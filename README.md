# gouzuru

a golang implementation of some functionality in https://github.com/carterjones/nouzuru

This is primarily an exercise in learning Go. If it turns out to be useful, I may continue development and extend the feature set further.

Note: only 64-bit systems running greater than or equal to Windows 7 are supported.

## Concurrency notes

Here's some totally not scientific data. I scanned every readable region in sublime_text.exe 100 times. The first time, I scanned it using sequential reads (calls to ReadProcessMemory) and the next time, I used goroutines for each call to ReadProcessMemory. I'm sure my implementation is poor in many ways, yet the results are still very interesting (here's three invocations of the test program). This was created using commit 46beb00:

    .\gouzuru.exe -p sublime_text.exe
    Successfully opened sublime_text.exe. PID: 7448. Handle: 620.
    Sequental read time:  7.8843326s
    Concurrent read time: 5.8090427s

    .\gouzuru.exe -p sublime_text.exe
    Successfully opened sublime_text.exe. PID: 7448. Handle: 596.
    Sequental read time:  9.8039849s
    Concurrent read time: 6.0332909s

    .\gouzuru.exe -p sublime_text.exe
    Successfully opened sublime_text.exe. PID: 7448. Handle: 584.
    Sequental read time:  9.5675209s
    Concurrent read time: 5.7121616s
