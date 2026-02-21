-- 删除已存在的表
DROP TABLE IF EXISTS `user`;

-- 创建 user 表
CREATE TABLE `user` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '用户姓名',
    `age` INT NOT NULL DEFAULT 0 COMMENT '年龄',
    `sex` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '性别',
    `score` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '分数',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_name` (`name`),
    KEY `idx_age` (`age`),
    KEY `idx_score` (`score`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 插入测试数据
INSERT INTO `user` (`name`, `age`, `sex`, `score`) VALUES
('张三', 25, '男', 85.50),
('李四', 30, '女', 92.00),
('王五', 22, '男', 78.75),
('赵六', 28, '女', 88.25),
('孙七', 35, '男', 95.00),
('周八', 26, '女', 81.50),
('吴九', 29, '男', 87.00),
('郑十', 24, '女', 90.25),
('钱十一', 31, '男', 83.75),
('陈十二', 27, '女', 89.50);

-- 查询验证
SELECT * FROM `user`;