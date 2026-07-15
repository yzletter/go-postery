-- Agent 微服务模块建表语句

-- 切到 go_postery 数据库
use go_postery;


CREATE TABLE IF NOT EXISTS interview_profiles
(
    id          BIGINT PRIMARY KEY,
    name        VARCHAR(256) DEFAULT '',
    skill_level JSON,
    weak_points JSON,
    created_at  DATETIME     DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS interview_records
(
    id               BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id          VARCHAR(128) NOT NULL,
    session_id       VARCHAR(128) NOT NULL UNIQUE,
    position         VARCHAR(256) DEFAULT '',
    overall_score    DOUBLE       DEFAULT 0,
    report_json      MEDIUMTEXT,
    review_plan_json MEDIUMTEXT,
    created_at       DATETIME     DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_session_id (session_id)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS interview_sessions
(
    id           VARCHAR(128) PRIMARY KEY,
    user_id      VARCHAR(128) NOT NULL,
    session_data MEDIUMTEXT,
    created_at   DATETIME    DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME    DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;