ffmpeg \
  -input_format mjpeg -video_size 1280x720 -framerate 30 -i /dev/video0 \
  -c:v libx264 -vf format=yuv420p -g 30 \
  -seg_duration 2 \
  -ldash 1 \
  -use_template 1 \
  -use_timeline 0 \
  -utc_timing_url "https://time.akamai.com/?iso" \
  -tune zerolatency \
  -streaming 1 \
  -method PUT \
  -http_persistent 1 \
  -f dash http://127.0.0.1:8080/ldash/stream.mpd

  -frag_duration 1 \
  -frag_type duration \
