#!/bin/bash
export DISPLAY=:0
while true; do
    echo "啟動應用程式..."
    APP_PATH_PLACEHOLDER
    echo "應用程式已關閉，1秒後重新啟動..."
    sleep 1
done
