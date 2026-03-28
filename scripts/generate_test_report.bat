@echo off
REM Gonest Framework Test Report Generator for Windows

echo ==========================================
echo   Gonest Framework Test Report Generator
echo ==========================================
echo.

REM Create reports directory
if not exist reports mkdir reports

REM Run tests with coverage
echo Running tests with coverage...
go test ./... -coverprofile=reports/coverage.out -covermode=atomic 2>&1 | tee reports/test_output.txt

REM Generate coverage report
echo.
echo Generating coverage report...
go tool cover -html=reports/coverage.out -o reports/coverage.html 2>nul

REM Count tests
find /c "PASS" reports\test_output.txt > reports\pass_count.txt 2>nul
find /c "FAIL" reports\test_output.txt > reports\fail_count.txt 2>nul

REM Run benchmarks
echo.
echo Running benchmarks...
go test -bench=. -benchmem -run=^$ ./... 2>&1 | tee reports/benchmark.txt

echo.
echo ==========================================
echo   Test Report Generated Successfully!
echo ==========================================
echo.
echo Reports generated in reports/ directory:
echo    - coverage.html : Coverage visualization
echo    - coverage.out  : Coverage data
echo    - benchmark.txt : Benchmark results
echo    - test_output.txt : Raw test output
echo.

pause