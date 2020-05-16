USE demo;
GO

DROP TABLE IF EXISTS PurchaseUsers;					
GO

CREATE TABLE PurchaseUsers (
    id         INT IDENTITY,
    name       VARCHAR(255) NOT NULL,
    create_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(id)
);
GO

DROP TABLE IF EXISTS ThirdpartyVendors;					
GO

CREATE TABLE ThirdpartyVendors (
    id         INT IDENTITY,
    name       VARCHAR(255) NOT NULL,
    vid        VARCHAR(255) NOT NULL,
    create_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(id)
);
GO