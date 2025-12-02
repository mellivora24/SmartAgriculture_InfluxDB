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
