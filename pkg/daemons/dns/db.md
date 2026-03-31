🧱 Core Tables

CREATE TABLE zones (
    id UUID,
    domain TEXT,
    namespace TEXT,
    enabled BOOLEAN,
    ns_primary TEXT,
    hostmaster TEXT,
    serial INT,
    refresh INT,
    retry INT,
    expire INT,
    minttl INT,
    soa_ttl INT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY ((domain), namespace)
);

CREATE TABLE active_zones (
    domain TEXT PRIMARY KEY,
    zone_id UUID,
    namespace TEXT
);

CREATE TABLE records (
    zone_id UUID,
    name TEXT,
    type SMALLINT,
    content TEXT,
    ttl INT,
    priority INT,
    weight INT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY ((zone_id), name, type, content)
);

CREATE TABLE active_records (
    fqdn TEXT,
    zone_id UUID,
    type SMALLINT,
    content TEXT,
    ttl INT,
    priority INT,
    weight INT,

    PRIMARY KEY ((fqdn, type), content)
);

CREATE TABLE zones_by_id (
    id UUID PRIMARY KEY,
    domain TEXT,
    namespace TEXT,
    ns_primary TEXT,
    hostmaster TEXT,
    serial INT,
    refresh INT,
    retry INT,
    expire INT,
    minttl INT,
    soa_ttl INT
);

🧱 Sample Data for fastpath lookup.

-- one zone id
-- 11111111-1111-1111-1111-111111111111

INSERT INTO active_records (fqdn, zone_id, type, content, ttl)
VALUES ('example.com.', 11111111-1111-1111-1111-111111111111, 1, '1.2.3.4', 300);

INSERT INTO active_records (fqdn, zone_id, type, content, ttl)
VALUES ('example.com.', 11111111-1111-1111-1111-111111111111, 28, '2001:db8::1', 300);

INSERT INTO active_records (fqdn, zone_id, type, content, ttl)
VALUES ('example.com.', 11111111-1111-1111-1111-111111111111, 16, 'hello-world', 300);

INSERT INTO active_records (fqdn, zone_id, type, content, ttl)
VALUES ('www.example.com.', 11111111-1111-1111-1111-111111111111, 5, 'example.com.', 300);

INSERT INTO active_records (fqdn, zone_id, type, content, ttl)
VALUES ('*.example.com.', 11111111-1111-1111-1111-111111111111, 1, '5.6.7.8', 300);
