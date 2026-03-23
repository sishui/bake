DROP TABLE IF EXISTS `test_all_types`;
CREATE TABLE `test_all_types`
(
    `id`               int          NOT NULL AUTO_INCREMENT,
    `tiny_int_val`     tinyint               DEFAULT NULL,
    `small_int_val`    smallint              DEFAULT 0 NOT NULL,
    `medium_int_val`   mediumint             DEFAULT NULL,
    `int_val`          int                   DEFAULT 0 NOT NULL,
    `big_int_val`      bigint                DEFAULT NULL,
    `decimal_val`      decimal(10, 2)        DEFAULT NULL,
    `float_val`        float                 DEFAULT NULL,
    `double_val` double DEFAULT NULL,
    `char_val`         char(10)              DEFAULT NULL,
    `varchar_val`      varchar(255)          DEFAULT NULL,
    `text_val`         text,
    `mediumtext_val`   mediumtext,
    `longtext_val`     longtext,
    `date_val`         date                  DEFAULT NULL,
    `time_val`         time                  DEFAULT NULL,
    `datetime_val`     datetime              DEFAULT NULL,
    `timestamp_val`    timestamp NULL DEFAULT NULL,
    `year_val` year DEFAULT NULL,
    `bool_val`         tinyint(1) DEFAULT NULL,
    `enum_val`         enum('A','B','C') DEFAULT NULL,
    `set_val` set('X','Y','Z') DEFAULT NULL,
    `binary_val`       binary(8) DEFAULT NULL,
    `varbinary_val`    varbinary(255) DEFAULT NULL,
    `blob_val`         blob,
    `mediumblob_val`   mediumblob,
    `longblob_val`     longblob,
    `json_val`         json                  DEFAULT NULL COMMENT 'This is a JSON column\nwith multiple lines of text\r\nThis is a JSON column\nwith multiple lines of text\r\nThis is a JSON column\nwith multiple lines of text',
    `bit_val`          bit(8)                DEFAULT NULL,
    `nullable_int_val` int                   DEFAULT NULL,
    `ignore_val`       text,
    `created_at`       TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'created at',
    `updated_at`       TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'updated at',
    `deleted_at`       TIMESTAMP(3) NULL DEFAULT NULL COMMENT 'deleted at',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB;


DROP TABLE IF EXISTS `mail_attachment`;
CREATE TABLE `mail_attachment`
(
    `id`           BIGINT  NOT NULL AUTO_INCREMENT COMMENT 'pk',
    `mail_id`      BIGINT  NOT NULL COMMENT 'email id',
    `kind`         TINYINT NOT NULL DEFAULT 0 COMMENT 'attachment kind',
    `reward_id`    BIGINT  NOT NULL COMMENT 'attachment reward id',
    `reward_count` INT UNSIGNED NOT NULL COMMENT 'attachment reward count (must > 0)',
    `created_at`   BIGINT  NOT NULL COMMENT 'created at',
    `updated_at`   BIGINT  NOT NULL COMMENT 'updated at',
    PRIMARY KEY (`id`),
    INDEX          `idx_mail_id` (`mail_id`)
) ENGINE=InnoDB COMMENT='user email attachment';

DROP TABLE IF EXISTS `mail`;
CREATE TABLE `mail`
(
    `id`             BIGINT       NOT NULL AUTO_INCREMENT COMMENT 'pk',
    `uid`            BIGINT       NOT NULL COMMENT 'user id',
    `subject`        VARCHAR(128) NOT NULL COMMENT 'email subject',
    `content`        TEXT         NOT NULL COMMENT 'email content',
    `status`         TINYINT      NOT NULL DEFAULT 0 COMMENT '0: unread, 1: read, 2: collected',
    `has_attachment` TINYINT      NOT NULL DEFAULT 0 COMMENT 'has attachment: 0: no, 1: yes',
    `created_at`     TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'created at',
    `updated_at`     TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'updated at',
    PRIMARY KEY (`id`),
    INDEX            `idx_uid_created` (`uid`, `created_at` DESC),
    INDEX            `idx_uid_status` (`uid`, `status`)
) ENGINE=InnoDB COMMENT='user email';

INSERT INTO `test_all_types` (tiny_int_val, small_int_val, medium_int_val, int_val, big_int_val,
                              decimal_val, float_val, double_val,
                              char_val, varchar_val, text_val, mediumtext_val, longtext_val,
                              date_val, time_val, datetime_val, timestamp_val, year_val,
                              bool_val, enum_val, set_val,
                              binary_val, varbinary_val,
                              blob_val, mediumblob_val, longblob_val,
                              json_val, bit_val,
                              nullable_int_val, ignore_val)
VALUES (1, 100, 1000, 10000, 100000,
        123.45, 1.23, 3.1415926,
        'char_test', 'varchar_test', 'text...', 'medium text...', 'long text...',
        '2024-01-01', '12:34:56', '2024-01-01 12:34:56', CURRENT_TIMESTAMP, 2024,
        1, 'A', 'X,Y',
        'abcdefgh', 'varbinary_data',
        'blobdata', 'mediumblobdata', 'longblobdata',
        JSON_OBJECT('key', 'value'),
        b'10101010',
        NULL, 'ignore me'),
       (-1, 0, NULL, 0, NULL,
        NULL, NULL, NULL,
        'abc', 'hello world', NULL, NULL, NULL,
        NULL, NULL, NULL, NULL, NULL,
        0, 'B', 'Z',
        '12345678', 'bin',
        NULL, NULL, NULL,
        JSON_ARRAY(1, 2, 3),
        b'00001111',
        42, NULL);


INSERT INTO `mail` (uid, subject, content, status, has_attachment)
VALUES (1001, 'Welcome', 'Welcome to the system', 0, 0),
       (1001, 'Reward Mail', 'You got rewards', 1, 1),
       (1002, 'System Notice', 'Maintenance notice', 0, 0);


INSERT INTO `mail_attachment` (mail_id, kind, reward_id, reward_count, created_at, updated_at)
VALUES (2, 1, 10001, 10, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
       (2, 2, 20001, 5, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
       (2, 1, 30001, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 外键示例: users + posts
DROP TABLE IF EXISTS `posts`;
DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(64)  NOT NULL COMMENT '用户名',
  `email`      VARCHAR(128) NOT NULL COMMENT '邮箱',
  `created_at` TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_email` (`email`)
) ENGINE=InnoDB COMMENT='用户表';

CREATE TABLE `posts` (
  `id`         BIGINT        NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT        NOT NULL COMMENT '用户ID',
  `title`      VARCHAR(256)  NOT NULL COMMENT '标题',
  `content`    TEXT          COMMENT '内容',
  `status`     TINYINT       NOT NULL DEFAULT 0 COMMENT '状态: 0=草稿, 1=发布',
  `created_at` TIMESTAMP(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` TIMESTAMP(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  CONSTRAINT `fk_posts_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB COMMENT='文章表';

INSERT INTO `users` (`name`, `email`) VALUES
  ('Alice', 'alice@example.com'),
  ('Bob', 'bob@example.com');

INSERT INTO `posts` (`user_id`, `title`, `content`, `status`) VALUES
  (1, 'First Post', 'Hello World!', 1),
  (1, 'Second Post', 'Another post by Alice', 0),
  (2, 'Bob Post', 'Post by Bob', 1);
