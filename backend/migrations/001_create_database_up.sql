CREATE TABLE tbl_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users_email ON tbl_users(email);
CREATE INDEX idx_users_username ON tbl_users(username);

CREATE TABLE tbl_farms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_farms_name ON tbl_farms(name);

CREATE TABLE tbl_farm_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES tbl_users(id) ON DELETE CASCADE,
    farm_id UUID NOT NULL REFERENCES tbl_farms(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'viewer', -- owner, manager, viewer
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, farm_id)
);

CREATE INDEX idx_farm_users_user_id ON tbl_farm_users(user_id);
CREATE INDEX idx_farm_users_farm_id ON tbl_farm_users(farm_id);

CREATE TABLE tbl_mcus (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    farm_id UUID NOT NULL REFERENCES tbl_farms(id) ON DELETE CASCADE,
    mcu_code VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'offline',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_mcus_farm_id ON tbl_mcus(farm_id);
CREATE INDEX idx_mcus_mcu_code ON tbl_mcus(mcu_code);
CREATE INDEX idx_mcus_status ON tbl_mcus(status);

CREATE TABLE tbl_survey_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mcu_id UUID NOT NULL REFERENCES tbl_mcus(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'connecting',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_survey_points_mcu_id ON tbl_survey_points(mcu_id);

CREATE TABLE tbl_device_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_point_id UUID REFERENCES tbl_survey_points(id) ON DELETE SET NULL,
    device_name VARCHAR(255) NOT NULL,
    command VARCHAR(50) NOT NULL, -- on, off
    status VARCHAR(50) DEFAULT 'pending', -- pending, sent, success, failed
    executed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_device_commands_status ON tbl_device_commands(status);
CREATE INDEX idx_device_commands_created_at ON tbl_device_commands(created_at);

CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON tbl_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_farms_updated_at BEFORE UPDATE ON tbl_farms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mcus_updated_at BEFORE UPDATE ON tbl_mcus
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_survey_points_updated_at BEFORE UPDATE ON tbl_survey_points
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE tbl_users IS 'Người dùng hệ thống';
COMMENT ON TABLE tbl_farms IS 'Nông trại/Trang trại';
COMMENT ON TABLE tbl_farm_users IS 'Quan hệ người dùng và nông trại';
COMMENT ON TABLE tbl_mcus IS 'MCU chính (ESP32 Gateway)';
COMMENT ON TABLE tbl_survey_points IS 'Điểm khảo sát';
COMMENT ON TABLE tbl_device_commands IS 'Lịch sử lệnh điều khiển';

-- New tbl
-- Bảng cấu hình ngưỡng cảnh báo cho từng survey point
CREATE TABLE tbl_threshold_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_point_id UUID NOT NULL REFERENCES tbl_survey_points(id) ON DELETE CASCADE,
    -- Ngưỡng nhiệt độ
    temp_min FLOAT,
    temp_max FLOAT,
    temp_critical_min FLOAT,
    temp_critical_max FLOAT,
    -- Ngưỡng độ ẩm không khí
    humidity_min FLOAT,
    humidity_max FLOAT,
    humidity_critical_min FLOAT,
    humidity_critical_max FLOAT,
    -- Ngưỡng độ ẩm đất
    soil_moisture_min FLOAT,
    soil_moisture_max FLOAT,
    soil_moisture_critical_min FLOAT,
    soil_moisture_critical_max FLOAT,
    -- Ngưỡng ánh sáng
    light_min FLOAT,
    light_max FLOAT,
    light_critical_min FLOAT,
    light_critical_max FLOAT,
    -- Cài đặt tự động bơm nước
    auto_pump_enabled BOOLEAN DEFAULT false,
    pump_trigger_soil_moisture FLOAT, -- Ngưỡng độ ẩm đất để bật máy bơm
    pump_stop_soil_moisture FLOAT,    -- Ngưỡng độ ẩm đất để tắt máy bơm
    pump_duration_seconds INT DEFAULT 30, -- Thời gian bơm tối đa (giây)
    pump_cooldown_minutes INT DEFAULT 60, -- Thời gian nghỉ giữa các lần bơm (phút)
    -- Cài đặt cảnh báo
    alert_enabled BOOLEAN DEFAULT true,
    alert_cooldown_minutes INT DEFAULT 10, -- Thời gian giữa các cảnh báo giống nhau
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(survey_point_id)
);

CREATE INDEX idx_threshold_settings_survey_point_id ON tbl_threshold_settings(survey_point_id);

-- Bảng lịch sử cảnh báo
CREATE TABLE tbl_alert_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_point_id UUID NOT NULL REFERENCES tbl_survey_points(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL, -- temperature, humidity, soil_moisture, light
    severity VARCHAR(20) NOT NULL, -- warning, critical
    sensor_value FLOAT NOT NULL,
    threshold_value FLOAT NOT NULL,
    message TEXT NOT NULL,
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by UUID REFERENCES tbl_users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alert_history_survey_point_id ON tbl_alert_history(survey_point_id);
CREATE INDEX idx_alert_history_created_at ON tbl_alert_history(created_at);
CREATE INDEX idx_alert_history_acknowledged ON tbl_alert_history(acknowledged);
CREATE INDEX idx_alert_history_severity ON tbl_alert_history(severity);

-- Bảng lịch sử tự động bơm
CREATE TABLE tbl_auto_pump_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_point_id UUID NOT NULL REFERENCES tbl_survey_points(id) ON DELETE CASCADE,
    command_id UUID REFERENCES tbl_device_commands(id),
    trigger_soil_moisture FLOAT NOT NULL,
    target_soil_moisture FLOAT NOT NULL,
    pump_duration_seconds INT NOT NULL,
    status VARCHAR(50) DEFAULT 'triggered', -- triggered, running, completed, failed
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT
);

CREATE INDEX idx_auto_pump_history_survey_point_id ON tbl_auto_pump_history(survey_point_id);
CREATE INDEX idx_auto_pump_history_started_at ON tbl_auto_pump_history(started_at);
CREATE INDEX idx_auto_pump_history_status ON tbl_auto_pump_history(status);

-- Trigger cho updated_at
CREATE TRIGGER update_threshold_settings_updated_at 
    BEFORE UPDATE ON tbl_threshold_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE tbl_threshold_settings IS 'Cấu hình ngưỡng cảnh báo và tự động bơm';
COMMENT ON TABLE tbl_alert_history IS 'Lịch sử các cảnh báo đã được gửi';
COMMENT ON TABLE tbl_auto_pump_history IS 'Lịch sử tự động bơm nước';

-- Insert default thresholds (example)
INSERT INTO tbl_threshold_settings (
    survey_point_id, 
    temp_min, temp_max, temp_critical_min, temp_critical_max,
    humidity_min, humidity_max, humidity_critical_min, humidity_critical_max,
    soil_moisture_min, soil_moisture_max, soil_moisture_critical_min, soil_moisture_critical_max,
    light_min, light_max, light_critical_min, light_critical_max,
    auto_pump_enabled, pump_trigger_soil_moisture, pump_stop_soil_moisture
) 
SELECT 
    id,
    15.0, 35.0, 10.0, 40.0,  -- Temperature
    30.0, 80.0, 20.0, 90.0,  -- Humidity
    30.0, 80.0, 20.0, 90.0,  -- Soil Moisture
    100.0, 50000.0, 50.0, 70000.0,  -- Light
    false, 30.0, 60.0  -- Auto pump settings
FROM tbl_survey_points
WHERE NOT EXISTS (
    SELECT 1 FROM tbl_threshold_settings WHERE survey_point_id = tbl_survey_points.id
);