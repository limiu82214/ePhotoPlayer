#!/bin/bash
# setup_auto_restart.sh
# 此腳本以冪等方式設定應用程式的自動重啟功能，
# 使用獨立的 auto_restart.sh 與 systemd 服務檔案進行設定。
# 並自動填入正確的使用者名稱。
# 請以 root 權限執行此腳本（例如使用 sudo）。

# 取得目前目錄的絕對路徑
CUR_DIR="$(pwd)"

# 判斷使用者：如果 SUDO_USER 存在則使用它，否則使用 whoami 的輸出
if [ -n "$SUDO_USER" ]; then
    USERNAME="$SUDO_USER"
else
    USERNAME="$(whoami)"
fi

# 定義路徑
APP_PATH="$CUR_DIR/ePhotoPlayer"              # 應用程式位於當前目錄下
SCRIPT_PATH="$CUR_DIR/auto_restart.sh"          # auto_restart.sh 應存在於當前目錄
SERVICE_SOURCE="$CUR_DIR/auto_restart.service"  # 當前目錄中的獨立服務檔案
SERVICE_DEST="/etc/systemd/system/auto_restart.service"  # 服務檔案目標位置

# 檢查 auto_restart.sh 是否存在
if [ ! -f "$SCRIPT_PATH" ]; then
  echo "錯誤：在當前目錄 ($CUR_DIR) 找不到 auto_restart.sh。"
  exit 1
fi

# 檢查 auto_restart.service 是否存在
if [ ! -f "$SERVICE_SOURCE" ]; then
  echo "錯誤：在當前目錄 ($CUR_DIR) 找不到 auto_restart.service。"
  exit 1
fi

echo "更新 auto_restart.sh..."
# 將 auto_restart.sh 中的 APP_PATH_PLACEHOLDER 替換為實際應用程式路徑
sed -i "s|APP_PATH_PLACEHOLDER|$APP_PATH|g" "$SCRIPT_PATH"
chmod +x "$SCRIPT_PATH"
echo "auto_restart.sh 已就緒於 $SCRIPT_PATH."

echo "更新 auto_restart.service..."
# 將服務檔案中的 SCRIPT_PATH_PLACEHOLDER 替換為 auto_restart.sh 的絕對路徑
sed -i "s|SCRIPT_PATH_PLACEHOLDER|$SCRIPT_PATH|g" "$SERVICE_SOURCE"
# 將服務檔案中的 USERNAME_PLACEHOLDER 替換為正確的使用者名稱
sed -i "s|USERNAME_PLACEHOLDER|$USERNAME|g" "$SERVICE_SOURCE"

echo "將 auto_restart.service 複製到 $SERVICE_DEST..."
# 複製更新後的服務檔案到 /etc/systemd/system/
cp "$SERVICE_SOURCE" "$SERVICE_DEST"

echo "重新載入 systemd 設定..."
systemctl daemon-reload

echo "啟用並重新啟動自動重啟服務..."
systemctl enable auto_restart.service
systemctl restart auto_restart.service

echo "設定完成，自動重啟服務已啟動，使用者：$USERNAME。"
systemctl status auto_restart.service
