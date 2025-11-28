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
    id          int auto_increment comment '用户 id, 自增',
    name        varchar(20) not null comment '用户名',
    password    char(32)    not null comment '用户密码的 md5 加密结果',
    create_time datetime default current_timestamp comment '用户注册时间, 默认为创建记录的时间',
    update_time datetime default current_timestamp on update current_timestamp comment '用户最后修改时间',
    primary key (id),
    unique key idx_name (name)
) default charset = utf8mb4 comment '用户信息表';


# 创建 post 表
create table if not exists post
(
    id          int auto_increment comment '帖子 id, 自增',
    user_id     int not null comment '发布者 id',
    create_time datetime default current_timestamp comment '帖子创建时间',
    update_time datetime default current_timestamp on update current_timestamp comment '帖子最后修改时间',
    delete_time datetime default null comment '帖子删除时间',
    title      varchar(100) comment '标题',
    content     text comment '正文',
    primary key (id),
    key idx_user (user_id)
) default charset = utf8mb4 comment '帖子信息表';