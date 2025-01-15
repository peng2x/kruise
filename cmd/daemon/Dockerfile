FROM mcr.microsoft.com/dotnet/runtime:8.0-windowsservercore-ltsc2019

WORKDIR /
COPY ./daemon.exe ./kruise-daemon.exe

ENTRYPOINT ["./kruise-daemon.exe"]
