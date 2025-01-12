DROP TABLE IF EXISTS Users;
DROP TABLE IF EXISTS RefreshTokens;

CREATE TABLE Users (
    ID int auto_increment,
    Username varchar(50) NOT NULL,
    Password varchar(255) DEFAULT NULL,
    PRIMARY KEY (ID)
);

CREATE TABLE RefreshTokens (
    ID varchar(255) NOT NULL,
    User_ID int,
    CreatedAt timestamp(4),
    ExpiresAt timestamp(4),
    PRIMARY KEY (ID),
    UNIQUE KEY(User_ID)
);