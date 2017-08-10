#!/usr/bin/env ruby

require 'bgp4r'
require 'json'
require 'optparse'
require 'packetfu'


neighbors = []
opts = OptionParser.new do |opts|
  opts.banner = 'bgp packet parser for debugging'
  opts.define_head "Usage: #$0 -1 local_address1 -2 local_address2 pcap_file"
  opts.separator ''
  opts.separator 'Options:'

  opts.on '-1 ADDR', 'Local address 1' do |v|
    neighbors << v
  end

  opts.on '-2 ADDR', 'Local address 2' do |v|
    neighbors << v
  end

  opts.on_tail '-h', '--help', 'Show this message' do
    puts opts
    exit
  end
end
opts.parse!(ARGV)

unless ARGV.size >= 1 && neighbors.size >= 2
  abort opts.to_s
end


class Parser
  TIMEOUT = 10
  SCENARIOS = %w(advertise_new_routes advertise_strong_routes withdraw_strong_routes withdraw_last_routes)

  def initialize(neighbors)
    @neighbors = neighbors
    @scenario = 0
    @result = [{time: nil, received: 0, advertised: 0}]
  end

  def parse(string)
    data = _parse(string)
    return unless data

    @tick ||= data[:time].to_i + 1

    if data[:time] > @tick
      @result << {time: nil, received: @result.last[:received], advertised: @result.last[:advertised]}
      @tick += 1
    end

    if data[:time] > @tick + TIMEOUT
      save

      @result = [{time: nil, received: 0, advertised: 0}]
      @tick = nil
      @scenario += 1
    end

    send SCENARIOS[@scenario], data
  end

  def _parse(string)
    string =~ /Epoch Time: (\S+) seconds.*Internet Protocol Version ., Src: (\S+), Dst: (\S+)/m
    data = {time: $1.to_i, src: $2, dst: $3}
    data[:nlri] = string.scan(/NLRI prefix:/).size
    data[:withdrawn] = string.scan(/Withdrawn prefix:/).size

    if data[:nlri] > 0 || data[:withdrawn] > 0
      data
    else
      nil
    end
  end

  def advertise_new_routes(data)
    @result.last[:time] ||= data[:time].to_i + 1

    if data[:src] == @neighbors[0]
      @result.last[:received] += data[:nlri]
    elsif data[:dst] == @neighbors[1]
      @result.last[:advertised] += data[:nlri]
    end
  end

  def advertise_strong_routes(data)
    @result.last[:time] ||= data[:time].to_i + 1

    if data[:src] == @neighbors[1]
      @result.last[:received] += data[:nlri]
    elsif data[:dst] == @neighbors[0]
      @result.last[:advertised] += data[:nlri]
    end
  end

  def withdraw_strong_routes(data)
    @result.last[:time] ||= data[:time].to_i + 1

    if data[:src] == @neighbors[1]
      @result.last[:received] += data[:withdrawn]
    elsif data[:dst] == @neighbors[0]
      @result.last[:advertised] += data[:nlri]
    end
  end

  def withdraw_last_routes(data)
    @result.last[:time] ||= data[:time].to_i + 1

    if data[:src] == @neighbors[0]
      @result.last[:received] += data[:withdrawn]
    elsif data[:dst] == @neighbors[1]
      @result.last[:advertised] += data[:withdrawn]
    end
  end

  def save
    @result.reject! {|i| i[:time].nil? }
    @result.each do |i|
      i[:time] = Time.at(i[:time]).strftime('%T')
    end

    Dir.mkdir('report') unless File.directory?('report')

    file = "report/#{SCENARIOS[@scenario]}.json"
    puts %(writing to "#{file}")

    open(file, 'w') do |f|
      f.write JSON.dump(@result)
    end
  end
end


parser = Parser.new(neighbors)
IO.popen("tshark -Vr #{ARGV.first}") do |io|
  while lines = io.gets("\n\n")
    parser.parse lines
  end
  parser.save
end
