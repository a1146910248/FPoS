@echo off
for /l %%i in (1,1,30) do (
    echo Running low1run1.exe, iteration %%i
    start /B low1run1.exe
)
echo Task completed.
pause
