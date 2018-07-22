# Commango

## What is this?
Commango is a Daemon to monitor a serial port for 3D printers. This is a module of MangoOS.

### Program Description
Commango is a tool for monitoring serial ports. I hated the way that other serial monitors would take over the port without sharing it with any other programs. This program will attach to a serial monitor and share programatic control over it with other programs using DBUS. Other programs will be able to command this program to take control of any Serial port and share the output, However Third party programs will alsobe able to read this information. 

### Program Rant
Primarily I plan to use this program with 3D printers. I have used Octoprint's plugin system, however I did not like that I was constrained to use python, and that I had to plan my programs to fit within Octoprint's runtime. This causes Development slowdowns as I was waiting for the entire server to restart instead of my one corner of the program. (My program made a 30 second server restart into a 1 minute 30 second restart.) Then since we were using python I could not find my mistakes until I was running the code. Which led to multiple restarts for small clerical errors.This also made automated error detection difficult as I could not run tests on modules.(I still need to get in the habit of writing tests.) 

## Features

### Made
None

### Not Made
- Port Monitor
- Daemon
- CLI tool
- Port Profiles
- Multi-Port-Control
- Configuration File
- DBUS access
