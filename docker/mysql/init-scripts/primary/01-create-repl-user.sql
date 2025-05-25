-- E:/Doer_xyz/go-common/docker/mysql/init-scripts/primary/01-init-primary.sql
-- 此脚本用于初始化主数据库 (doer_mysql_primary)
-- 它将创建必要的数据库和复制用户

-- 创建 doer_userHub 数据库 (供 user-hub 服务使用)
CREATE DATABASE IF NOT EXISTS `doer_userHub`
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

-- 创建 doer_post_service 数据库 (供 post-service 服务使用)
CREATE DATABASE IF NOT EXISTS `doer_post_service`
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

-- (如果还有其他服务需要在此 MySQL 实例中创建独立数据库，请在此处添加)
-- CREATE DATABASE IF NOT EXISTS `another_service_db`
--     CHARACTER SET utf8mb4
--     COLLATE utf8mb4_unicode_ci;

-- 创建用于主从复制的用户 'repl_user'
-- 密码为 'repl_pass' (与 docker-compose.yaml 和从库配置脚本中的设置一致)
-- 重要: 使用 mysql_native_password 认证插件以避免 'caching_sha2_password' 导致的连接问题
CREATE USER IF NOT EXISTS 'repl_user'@'%' IDENTIFIED WITH mysql_native_password BY 'repl_pass';

-- 授予 'repl_user' REPLICATION SLAVE 权限，允许它从任何主机 (%) 连接
GRANT REPLICATION SLAVE ON *.* TO 'repl_user'@'%';

-- 刷新权限使更改生效
FLUSH PRIVILEGES;

-- (可选) 显示已创建的数据库，方便验证
-- SHOW DATABASES;