CREATE TABLE `roles` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL UNIQUE,
    `policies` JSON,
    PRIMARY KEY (`id`)
);

CREATE TABLE `users` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(255) NOT NULL UNIQUE,
    `password_hash` VARCHAR(255) NOT NULL,
    `role_id` INT,
       `created_at` TIMESTAMP,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`)
   );

CREATE TABLE `nodes` (
    `id` VARCHAR(255) NOT NULL,
    `hostname` VARCHAR(255),
    `status` VARCHAR(255),
    `gpus` JSON,
    `control_port` INT,
    `last_seen` TIMESTAMP,
    PRIMARY KEY (`id`)
);

CREATE TABLE `gpu_claims` (
    `id` VARCHAR(255) NOT NULL,
    `user_id` VARCHAR(255),
    `created_at` TIMESTAMP,
    `spec` JSON,
    `status` JSON,
    PRIMARY KEY (`id`)
   );