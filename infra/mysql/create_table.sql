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
create table if not exists user
(
    id            bigint      not null comment '用户 ID',
    name          varchar(20) not null comment '用户名',
    password      char(32)    not null comment '用户密码的 MD5 加密结果',
    email         varchar(128) comment '用户邮箱',
    avatar        varchar(255) comment '用户头像 URL',
    bio           varchar(255) comment '用户个性签名',
    gender        tinyint comment '用户性别',
    birthday      date comment '用户生日',
    location      varchar(64) comment '用户地区',
    country       varchar(64) comment '用户国家',
    status        tinyint comment '用户状态',
    last_login_ip varchar(64) comment '最近一次登录 IP',
    create_time   datetime default current_timestamp comment '用户注册时间, 默认为创建数据库记录的时间',
    update_time   datetime default current_timestamp on update current_timestamp comment '数据库记录最后修改时间',
    primary key (id),
    unique key idx_name (name),
    unique key idx_email (email)
) default charset = utf8mb4 comment '用户信息主表';

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
    name        varchar(32) not null comment '标签名',
    slug        varchar(32) not null comment '标签唯一标识',
    create_time datetime default current_timestamp comment '创建时间',
    delete_time datetime default null comment '删除时间',
    unique key uq_slug (slug),
    unique key uq_name (name)
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