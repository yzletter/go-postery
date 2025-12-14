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
    id            BIGINT UNSIGNED                         NOT NULL COMMENT '用户 ID (雪花算法)',
    username      VARCHAR(32) COLLATE utf8mb4_unicode_ci  NOT NULL COMMENT '用户名',
    email         VARCHAR(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮箱',
    password_hash VARCHAR(255)                            NOT NULL COMMENT '密码哈希',

    avatar        VARCHAR(255)                                     DEFAULT NULL COMMENT '头像 URL',
    bio           VARCHAR(255)                                     DEFAULT NULL COMMENT '个性签名',
    gender        TINYINT UNSIGNED                        NOT NULL DEFAULT 0 COMMENT '性别 0 空, 1 男, 2 女, 3 其他',
    birthday      DATE                                             DEFAULT NULL COMMENT '生日',
    location      VARCHAR(64)                                      DEFAULT NULL COMMENT '地区',
    country       VARCHAR(64)                                      DEFAULT NULL COMMENT '国家',
    status        TINYINT UNSIGNED                        NOT NULL DEFAULT 1 COMMENT '用户状态 1 正常, 2 封禁, 3 注销',
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
create table if not exists post
(
    id            bigint       not null comment '帖子 ID',
    user_id       bigint       not null comment '发布者 ID',
    title         varchar(255) not null comment '标题',
    view_count    int unsigned not null default 0 comment '浏览量',
    like_count    int unsigned not null default 0 comment '点赞数',
    comment_count int unsigned not null default 0 comment '评论数',

    status        tinyint               default 1 comment '状态',

    create_time   datetime              default current_timestamp comment '帖子创建时间',
    update_time   datetime              default current_timestamp on update current_timestamp comment '帖子最后修改时间',
    delete_time   datetime              default null comment '帖子删除时间',
    content       text comment '正文',
    primary key (id),
    unique key idx_user (user_id)
) default charset = utf8mb4 comment '帖子信息表';

use go_postery;
alter table post
    add column comment_count int unsigned not null default 0 comment '评论数';

create table if not exists user_like
(
    id          bigint auto_increment primary key comment '记录 ID',
    post_id     bigint not null comment '被点赞帖子 id',
    user_id     bigint not null comment '点赞者 id',
    create_time datetime default current_timestamp comment '帖子创建时间',
    update_time datetime default current_timestamp on update current_timestamp comment '帖子最后修改时间',
    delete_time datetime default null comment '帖子删除时间',
    unique key uq_user_post (user_id, post_id),
    key idx_target (post_id),
    KEY idx_user (user_id)
) default charset = utf8mb4 comment '用户点赞表';

create table if not exists comment
(
    id          bigint comment '评论 id',
    post_id     bigint not null comment '所属帖子 id',
    user_id     bigint not null comment '发布者 id',
    parent_id   bigint not null comment '父评论 id',
    reply_id    bigint not null comment '回复评论 id',
    create_time datetime default current_timestamp comment '帖子创建时间',
    delete_time datetime default null comment '帖子删除时间',
    content     text comment '正文',
    primary key (id),
    key idx_user (user_id)
) default charset = utf8mb4 comment '帖子信息表';


create table if not exists tag
(
    id          bigint primary key comment '标签 id',
    username        varchar(32) not null comment '标签名',
    slug        varchar(32) not null comment '标签唯一标识',
    create_time datetime default current_timestamp comment '创建时间',
    delete_time datetime default null comment '删除时间',
    unique key uq_slug (slug),
    unique key uq_name (username)
) default charset = utf8mb4 comment '标签信息表';

create table if not exists post_tag
(
    id      bigint primary key comment '记录 id',
    post_id bigint not null comment '帖子 id',
    tag_id  bigint not null comment '标签 id',
    unique key uq_post_tag (post_id, tag_id),
    key idx_tag (tag_id),
    key idx_post (post_id)
) default charset = utf8mb4 comment '帖子——标签绑定信息表';

create table if not exists follow
(
    id          bigint not null primary key comment '记录 id',
    follower_id bigint not null comment '关注者 id',
    followee_id bigint not null comment '被关注者 id',
    create_time datetime default current_timestamp comment '创建时间',
    delete_time datetime default null comment '删除时间',
    update_time datetime default current_timestamp on update current_timestamp comment '更新时间',

    unique key uq_follow (follower_id, followee_id),
    key idx_follower (follower_id, delete_time),
    key idx_followee (followee_id, delete_time),

    check (follower_id <> followee_id) # 避免自己关注自己
) default charset = utf8mb4 comment '关注信息表';