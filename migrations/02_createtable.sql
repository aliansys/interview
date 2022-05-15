create table events
(
    device_id    UUID,     -- "0287D9AA-4ADF-4B37-A60F-3E9E645C821E"
    device_os    String,   -- "iOS 13.5.1"
    session      String,   -- "ybuRi8mAUypxjbxQ"
    event        String,   -- "app_start"
    sequence     Int64,    -- 1
    param_int    Int64,    -- 0
    param_str    String,   -- "some text"
    ip           IPv4,    -- "8.8.8.8"
    client_time  DateTime, -- "2020-12-01 23:59:00"
    server_time  DateTime  -- "2020-12-01 23:53:00"
)
engine = MergeTree() PARTITION BY toYYYYMM(server_time) ORDER BY (device_id, event) SETTINGS index_granularity = 8192;