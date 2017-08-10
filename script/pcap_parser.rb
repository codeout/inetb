#!/usr/bin/env ruby

require 'bgp4r'
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


module Bgp
  MARKER = ["ffffffffffffffffffffffffffffffff"].pack('H*')

  class UpdateCounter
    attr_reader :result

    def initialize(neighbors)
      @neighbors = neighbors
      @prefixes = Hash.new{|h, k| h[k] = 0 }
      @result = {}
    end

    def read(raw_packet)
      packet = PacketFu::Packet.parse(raw_packet.data)
      timestamp = raw_packet.timestamp.sec.to_i
      payload = packet.payload

      return unless is_bgp?(payload)

      messages = payload.split(Bgp::MARKER)
      messages.shift  # BGP payload starts with themarker
      messages.each do |message|
        m = BGP::Message.factory(Bgp::MARKER + message)
        count(m, packet.ip_src_readable, packet.ip_dst_readable, timestamp)
      end
    end

    def count(message, src, dst, timestamp)
      return unless is_update?(message)

      unless @timestamp
        @timestamp = @start = timestamp
      end

      if timestamp > @timestamp
        @result[timestamp-@start] = @prefixes

        @prefixes = Hash.new{|h, k| h[k] = 0 }
        @timestamp = timestamp
      end

      key = key(src, dst)
      count_up key, message.nlri
      count_up key, message.withdrawn
    end


    private

    def is_bgp?(payload)
      payload.size >= 16 && !payload.start_with?(Bgp::MARKER)
    end

    def is_update?(message)
      message.class == BGP::Update
    end

    def key(src, dst)
      if @neighbors.include?(src)
        "#{dst} <- #{src}"
      else
        "#{src} -> #{dst}"
      end
    end

    def count_up(key, field)
      return unless field
      @prefixes[key] += field.nlris.size
    end
  end
end


counter = Bgp::UpdateCounter.new(neighbors)

PacketFu::PcapPackets.new.read(File.read(ARGV[0])).each do |packet|
  begin
    counter.read(packet)
  rescue
    $stderr.puts $!
  end
end

p counter.result
