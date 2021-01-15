# Aggregate loadtest results then output markdown table
# 
# Usage  : ruby aggregate-test-result-table.rb directory-of-result
# Example: ruby aggregate-test-result-table.rb 20201229-on-macbook/result
#
# Expected directory structure:
#   directory-of-result/
#     /channels-[0-9]+
#       /result
#         /summary.json   <- k6 JSON output file
require "json"

SUBSCRIBERS_PER_CHANNEL = 3

def main
    puts "| # | msg/sec | HTTP req/sec | Msg delay [ms]<br>(incl. 20ms sleep \`*1\`) | Publish API TTFB [ms] | Acknowledge API TTFB [ms] |"
    puts "| - | ------- | ------------ | --------------------------------------- | --------------------- | ------------------------- |"
    Dir.glob("#{ARGV[0]}/channels-*/result/summary.json").map{|json_file|
        {
            channels: File.basename(File.dirname(File.dirname(json_file))).match(/.+-([0-9]+)$/).captures[0].to_i,
            json: JSON.parse(File.read(json_file), symbolize_names: true),
        }
    }.sort_by{|map| map[:channels]}.each do |map|
        channels = map[:channels]
        json = map[:json]
        metrics = json[:metrics]
        columns = [
            "\`#{channels}\` (\`#{channels * SUBSCRIBERS_PER_CHANNEL}\`)",
            metrics[:dsps_fetched_messages][:rate],
            metrics[:http_reqs][:rate],
            format_gauge(metrics[:dsps_msg_delay_ms]),
            format_gauge(metrics[:dsps_ttfb_ms_publish]),
            format_gauge(metrics[:dsps_ttfb_ms_ack]),
        ]
        puts("| " + columns.map{|column|
            next "\`#{column.round(1)}\`" if column.is_a?(Numeric)
            next column.to_s
        }.join(" | ") + " |")
    end
end

def format_gauge(metric)
    # Values of 3-tuple are `median, 90 percentile, 95 percentile`.
    "\`" + %i(med p(90) p(95)).map{|aggregation|
        metric[aggregation].round(1).to_s
    }.join(", ") + "\`"
end

main()
