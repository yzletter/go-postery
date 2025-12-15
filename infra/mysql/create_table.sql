# 建库
-- 创建数据库 go_postery
create database go_postery;
-- 创建用户 go_postery_tester 密码为 123456
create user 'go_postery_tester' identified by '123456';
-- 将数据库 go_postery 的全部权限授予用户 go_postery_tester
grant all on go_postery.* to go_postery_tester;
-- 切到 go_postery 数据库
use go_postery;

# 建表
# 创建 user 表
CREATE TABLE IF NOT EXISTS users
(
    id            BIGINT                                  NOT NULL COMMENT '用户 ID (雪花算法)',
    username      VARCHAR(32) COLLATE utf8mb4_unicode_ci  NOT NULL COMMENT '用户名',
    email         VARCHAR(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮箱',
    password_hash VARCHAR(255)                            NOT NULL COMMENT '密码哈希',

    avatar        VARCHAR(255)                                     DEFAULT NULL COMMENT '头像 URL',
    bio           VARCHAR(255)                                     DEFAULT NULL COMMENT '个性签名',
    gender        TINYINT                                 NOT NULL DEFAULT 0 COMMENT '性别 0 空, 1 男, 2 女, 3 其他',
    birthday      DATE                                             DEFAULT NULL COMMENT '生日',
    location      VARCHAR(64)                                      DEFAULT NULL COMMENT '地区',
    country       VARCHAR(64)                                      DEFAULT NULL COMMENT '国家',
    status        TINYINT                                 NOT NULL DEFAULT 1 COMMENT '用户状态 1 正常, 2 封禁, 3 注销',
    last_login_ip VARCHAR(45)                                      DEFAULT NULL COMMENT '最后登录 IP',
    last_login_at DATETIME                                         DEFAULT NULL COMMENT '最后登录时间',
    created_at    DATETIME                                NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at    DATETIME                                NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at    DATETIME                                         DEFAULT NULL COMMENT '逻辑删除时间',

    PRIMARY KEY (id),

    UNIQUE KEY uk_user_username (username),
    UNIQUE KEY uk_user_email (email),

    KEY idx_user_status_deleted (status, deleted_at),

    CHECK (gender IN (0, 1, 2, 3)),
    CHECK (status IN (1, 2, 3))
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT '用户表';


# 创建 post 表
CREATE TABLE IF NOT EXISTS posts
(
    id            BIGINT       NOT NULL COMMENT '帖子 ID',
    user_id       BIGINT       NOT NULL COMMENT '发布者 ID',
    title         varchar(255) NOT NULL COMMENT '标题',
    content       TEXT COMMENT '正文',
    status        TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 1 正常, 2 封禁',
    view_count    INT          NOT NULL DEFAULT 0 COMMENT '浏览量',
    like_count    INT          NOT NULL DEFAULT 0 COMMENT '点赞数',
    comment_count INT          NOT NULL DEFAULT 0 COMMENT '评论数',

    created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at    DATETIME              DEFAULT NULL COMMENT '逻辑删除时间',

    PRIMARY KEY (id),
    KEY idx_user_created (user_id, created_at DESC),
    KEY idx_created (created_at DESC),
    KEY idx_status_deleted_created (status, deleted_at, created_at DESC)
) DEFAULT CHARSET = utf8mb4 COMMENT '帖子信息表';

# 创建 follow 表
CREATE TABLE IF NOT EXISTS follows
(
    id          BIGINT   NOT NULL PRIMARY KEY COMMENT '记录 id',
    follower_id BIGINT   NOT NULL COMMENT '关注者 id',
    followee_id BIGINT   NOT NULL COMMENT '被关注者 id',

    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at  DATETIME          DEFAULT NULL COMMENT '逻辑删除时间',

    UNIQUE KEY uq_follow (follower_id, followee_id),
    KEY idx_follower (follower_id, deleted_at),
    KEY idx_followee (followee_id, deleted_at),

    CHECK (follower_id <> followee_id) # 避免自己关注自己
) DEFAULT CHARSET = utf8mb4 COMMENT '关注信息表';

CREATE TABLE IF NOT EXISTS comments
(
    id         BIGINT   NOT NULL COMMENT '评论 id',
    post_id    BIGINT   NOT NULL COMMENT '所属帖子 id',
    user_id    BIGINT   NOT NULL COMMENT '发布者 id',
    parent_id  BIGINT   NOT NULL DEFAULT 0 COMMENT '父评论 id',
    reply_id   BIGINT   NOT NULL DEFAULT 0 COMMENT '回复评论 id',
    content    TEXT     NOT NULL COMMENT '正文',

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME          DEFAULT NULL COMMENT '逻辑删除时间',

    PRIMARY KEY (id),
    KEY idx_post_created (post_id, created_at),
    KEY idx_post_parent_created (post_id, parent_id, created_at),
    KEY idx_post_reply_created (post_id, reply_id, created_at)
) DEFAULT CHARSET = utf8mb4 COMMENT '评论信息表';


CREATE TABLE IF NOT EXISTS likes
(
    id         BIGINT COMMENT '记录 ID',
    post_id    BIGINT   NOT NULL COMMENT '被点赞帖子 id',
    user_id    BIGINT   NOT NULL COMMENT '点赞者 id',

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME          DEFAULT NULL COMMENT '逻辑删除时间',

    PRIMARY KEY (id),
    UNIQUE KEY uq_user_post (user_id, post_id),
    KEY idx_target (post_id),
    KEY idx_user (user_id),
    KEY idx_post_deleted (post_id, deleted_at)
) DEFAULT CHARSET = utf8mb4 COMMENT '用户点赞表';


CREATE TABLE IF NOT EXISTS tag
(
    id          BIGINT PRIMARY KEY COMMENT '标签 id',
    username    varchar(32) NOT NULL COMMENT '标签名',
    slug        varchar(32) NOT NULL COMMENT '标签唯一标识',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    delete_time DATETIME DEFAULT NULL COMMENT '删除时间',
    UNIQUE KEY uq_slug (slug),
    UNIQUE KEY uq_name (username)
) DEFAULT CHARSET = utf8mb4 COMMENT '标签信息表';

CREATE TABLE IF NOT EXISTS post_tag
(
    id      BIGINT PRIMARY KEY COMMENT '记录 id',
    post_id BIGINT NOT NULL COMMENT '帖子 id',
    tag_id  BIGINT NOT NULL COMMENT '标签 id',
    UNIQUE KEY uq_post_tag (post_id, tag_id),
    KEY idx_tag (tag_id),
    KEY idx_post (post_id)
) DEFAULT CHARSET = utf8mb4 COMMENT '帖子——标签绑定信息表';

