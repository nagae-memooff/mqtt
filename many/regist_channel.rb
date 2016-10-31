#!/usr/bin/env ruby
#encoding=utf-8
require 'redis'
# redis = Redis.new(:url => "redis://192.168.100.231:6379")
redis = Redis.new(:url => "redis://:abcd@192.168.100.228:6379")
redis.set "channel:fake_teding_recv", "teding_topic"

count = ARGV.first.to_i || 100
i = 0

size = 50000
j = 0
h = 0

while i < count do
  conf = ""
  redis.multi do
    while j < size && i < count do
    	c = "fake_recv_#{i}"
      key = "channel:#{c}"
      value = "topicabcdefg:#{c}"
      redis.set key, value
      conf += "#{c}=/u/#{value}\n"
      i += 1
      j += 1
    end
  end
  puts "#{h} finished."
  
  j = 0
  File.open "channel.list.#{h}", "w" do |f|
    f.write conf
  end
  h += 1
end
