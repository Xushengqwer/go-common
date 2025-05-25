#!/bin/bash
# E:/Doer_xyz/go-common/docker/mysql/init-scripts/replica/01-configure-replication.sh
# 此脚本用于配置从数据库 (doer_mysql_replica) 连接到主数据库并启动复制

set -e # 如果任何命令失败，则立即退出

# --- 配置变量 (与统一 docker-compose.yaml 保持一致) ---
PRIMARY_HOST="doer_mysql_primary" # 主库的服务名
PRIMARY_PORT="3306"               # 主库的内部端口
REPL_USER="repl_user"             # 复制用户名
REPL_PASS="repl_pass"             # 复制用户密码
ROOT_PASS="root"                  # 从库自身的 root 密码 (用于执行 CHANGE MASTER)
# --- 配置变量结束 ---

echo "Replica: Waiting for primary MySQL (${PRIMARY_HOST}:${PRIMARY_PORT}) to become available..."

# 循环等待主库可用
# -h"$PRIMARY_HOST": 主库服务名
# -P"$PRIMARY_PORT": 主库端口
# -uroot -p"$ROOT_PASS": 使用 root 用户和密码检查主库连接 (假设主库的 root 密码也是 "root")
# 您也可以在主库上创建一个低权限用户专门用于此健康检查
until mysqladmin ping -h"$PRIMARY_HOST" -P"$PRIMARY_PORT" -uroot -p"$ROOT_PASS" --silent; do
    echo "Replica: Primary (${PRIMARY_HOST}) is unavailable - sleeping for 3 seconds..."
    sleep 3
done

echo "Replica: Primary MySQL (${PRIMARY_HOST}) is available. Configuring replication..."

# 执行配置复制的 SQL 命令
# 使用从库的 root 用户和密码 (-p"${ROOT_PASS}")
mysql -uroot -p"${ROOT_PASS}" <<-EOSQL
    -- 停止当前可能正在运行的复制进程
    STOP REPLICA;

    -- 重置所有复制状态，清除旧的主库信息
    -- 对于全新设置或需要完全重置复制的场景是安全的
    RESET REPLICA ALL;

    -- 配置主库信息
    -- MASTER_AUTO_POSITION=1: 使用 GTID (全局事务标识符) 自动定位复制点
    -- 这要求主库已启用 GTID (在我们的 docker-compose.yaml 中已通过 command 参数为 mysql-primary 配置)
    CHANGE MASTER TO
        MASTER_HOST='${PRIMARY_HOST}',
        MASTER_USER='${REPL_USER}',
        MASTER_PASSWORD='${REPL_PASS}',
        MASTER_PORT=${PRIMARY_PORT},
        MASTER_AUTO_POSITION=1; -- 使用 GTID

    -- 启动复制进程
    START REPLICA;

    -- (可选) 在日志中显示从库状态，方便调试
    -- SHOW REPLICA STATUS\G;
EOSQL

echo "Replica: Replication setup script finished successfully."