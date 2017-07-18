#!/usr/bin/env ruby

require 'optparse'


output = 'chart.html'
reports = ['advertise_new_routes', 'advertise_strong_routes', 'withdraw_strong_routes', 'withdraw_last_routes']

opts = OptionParser.new do |opts|
  opts.banner = 'chart generator for inetb'
  opts.define_head 'Usage: make.rb path_to_report_dir'

  opts.on_tail '-h', '--help', 'Show this message' do
    puts opts
    exit
  end
end
opts.parse!

unless ARGV[0]
  puts opts
  abort
else
  report_dir = ARGV[0]
end

definition = ''
reports.each do |name|
  definition += %(  data['#{name}'] = #{File.read(File.join(report_dir, "#{name}.json")).strip};\n)
end

File.write File.join(report_dir, output), DATA.read.sub('__DATA__', definition)


__END__
<html>
  <head>
    <style>
     body { padding: 50px; }
    </style>
  </head>
  <body>
    <h2>Test Case 1: Receive New Prefixes via Peer 1</h2>
    <div id="advertise_new_routes" style="width: 900px; height: 500px"></div>

    <h2>Test Case 2: Receive Stronger Prefixes via Peer 2</h2>
    <div id="advertise_strong_routes" style="width: 900px; height: 500px"></div>

    <h2>Test Case 3: Receive Withdrawals for the Stronger Prefixes via Peer 2</h2>
    <div id="withdraw_strong_routes" style="width: 900px; height: 500px"></div>

    <h2>Test Case 4: Receive Withdrawals for the Rest of Prefixes via Peer 1</h2>
    <div id="withdraw_last_routes" style="width: 900px; height: 500px"></div>
  </body>

  <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
  <script type="text/javascript">
   google.charts.load('current', {packages: ['corechart', 'line']});
   google.charts.setOnLoadCallback(function () {
     draw('advertise_new_routes', 'Advertised Prefixes', 'Received Prefixes');
     draw('advertise_strong_routes', 'Advertised Prefixes', 'Received Prefixes');
     draw('withdraw_strong_routes', 'Advertised Withdrawn Prefixes', 'Received Withdrawn Prefixes');
     draw('withdraw_last_routes', 'Advertised Withdrawn Prefixes', 'Received Withdrawn Prefixes');
   });

  data = {};
__DATA__

   String.prototype.timeOfDay = function () {
     return  this.split(':').map(function(i) {
       return parseInt(i);
     });
   }

   function draw(name, advertisedLabel, receivedLabel) {
     var table = new google.visualization.DataTable();
     table.addColumn('timeofday', 'Time');
     table.addColumn('number', advertisedLabel)
     table.addColumn('number', receivedLabel)

     table.addRows(
       data[name].map(function(i) {
         return [i.time.timeOfDay(), i.advertised, i.received]
       })
     );

     var options = {
       hAxis: {
         title: 'Time'
       },
       vAxis: {
         title: 'Prefixes'
       }
     };

     var chart = new google.visualization.LineChart(document.getElementById(name));
     chart.draw(table, options);
   }
  </script>
</html>
