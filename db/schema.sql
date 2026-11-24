
CREATE DATABASE IF NOT EXISTS train_ticket_app
    CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

USE train_ticket_app;

CREATE TABLE IF NOT EXISTS trains (
                                      id VARCHAR(32) NOT NULL PRIMARY KEY,
                                      from_station VARCHAR(64) NOT NULL,
                                      to_station VARCHAR(64) NOT NULL,
                                      price_cents BIGINT NOT NULL,
                                      total_seats INT NOT NULL,
                                      available_seats INT NOT NULL,

                                      created_at DATETIME(3) NOT NULL
                                          DEFAULT CURRENT_TIMESTAMP(3),
                                      updated_at DATETIME(3) NOT NULL
                                          DEFAULT CURRENT_TIMESTAMP(3)
                                          ON UPDATE CURRENT_TIMESTAMP(3),

                                      CONSTRAINT ck_train_price CHECK (price_cents > 0),
                                      CONSTRAINT ck_train_total CHECK (total_seats > 0),
                                      CONSTRAINT ck_train_available CHECK (
                                          available_seats >= 0
                                              AND available_seats <= total_seats
                                          ),

                                      INDEX idx_train_route (from_station, to_station)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS purchase_logs (
                                             id BIGINT UNSIGNED NOT NULL
                                                 AUTO_INCREMENT PRIMARY KEY,
                                             train_id VARCHAR(32) NOT NULL,
                                             quantity INT NOT NULL,

                                             created_at DATETIME(3) NOT NULL
                                                 DEFAULT CURRENT_TIMESTAMP(3),

                                             CONSTRAINT ck_log_quantity CHECK (quantity > 0),

                                             CONSTRAINT fk_log_train FOREIGN KEY (train_id)
                                                 REFERENCES trains(id) ON DELETE RESTRICT,

                                             INDEX idx_log_train (train_id)
) ENGINE=InnoDB;
