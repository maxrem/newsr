CREATE TABLE `article`
(
    `id`          INT unsigned NOT NULL AUTO_INCREMENT,
    `title`       TEXT,
    `url`         VARCHAR(256),
    `hash`        BIGINT UNSIGNED UNIQUE,
    `description` TEXT,
    `content`     TEXT,
    `published`   DATETIME,
    PRIMARY KEY (`id`)
);