@echo off
:start
set /p a=insert manga id(0 to end): 
if %a%==0 goto exit
manzip.exe %a%
echo.
echo.
goto start
:exit
break