create table events
(
    device_id    UUID,
    device_os    String,
    session      String,
    event        String,
    sequence     Int64,
    param_int    Int64,
    param_str    String,
    ip           IPv4,
    client_time  DateTime,
    server_time  DateTime
)
engine = MergeTree() PARTITION BY toYYYYMM(server_time) ORDER BY (device_id, event) SETTINGS index_granularity = 8192;