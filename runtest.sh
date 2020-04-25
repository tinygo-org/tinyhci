# open serial port
stty -F /dev/ttyACM0 115200 raw -echo

exec 3</dev/ttyACM0                     # redirect serial to fd3
  cat <&3 > /tmp/testoutput.txt &       # redirect fd3 serial to file
  PID=$!                                # save pid to kill test if needed
    echo "t" > /dev/ttyACM0             # send command to start tests
    sleep 5.0s                          # wait for response
  kill $PID                             # kill test if takes too long
  wait $PID 2>/dev/null                 # surpress "Terminated" output

exec 3<&-                               # free fd3
cat /tmp/testoutput.txt                 # display test results
