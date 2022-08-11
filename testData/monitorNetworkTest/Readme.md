# Monitor Network Configuration

Each monitor number is connected to its corresponding gossiper number

Monitors are responsible for the "FakeCAs" and "FakeLoggers" (see those folders for info) as follows:

* 1 - logger1,logger2,CA1
* 2 - logger2, logger3, CA2
* 3 - logger1, logger3, CA1, CA3
* 4 - logger2, CA2, CA3

### Running a monitor
Go to the root of this project and run `sh ./monitorTest.sh N` where N is the number of the monitor you want to run. This runs ctng.go with the parameters from these test files. 

Then, in a seperate window, run `sh ./monitorTest.sh N` with the same N to launch the corresponding Monitor.