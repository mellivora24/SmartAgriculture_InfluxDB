-- ============================================
-- FUNCTION: get_user_farms
-- Lấy danh sách farms của một user
-- ============================================
CREATE OR REPLACE FUNCTION get_user_farms(p_user_id UUID)
    RETURNS TABLE (
                      farm_id UUID,
                      farm_name VARCHAR,
                      farm_description TEXT,
                      farm_location VARCHAR,
                      user_role VARCHAR,
                      mcu_count BIGINT,
                      survey_point_count BIGINT,
                      online_mcu_count BIGINT,
                      created_at TIMESTAMP WITH TIME ZONE
                  ) AS $$
BEGIN
    RETURN QUERY
        SELECT
            f.id,
            f.name,
            f.description,
            f.location,
            fu.role,
            COUNT(DISTINCT m.id) as mcu_count,
            COUNT(DISTINCT sp.id) as survey_point_count,
            COUNT(DISTINCT CASE WHEN m.status = 'online' THEN m.id END) as online_mcu_count,
            f.created_at
        FROM tbl_farms f
                 INNER JOIN tbl_farm_users fu ON f.id = fu.farm_id
                 LEFT JOIN tbl_mcus m ON f.id = m.farm_id
                 LEFT JOIN tbl_survey_points sp ON m.id = sp.mcu_id
        WHERE fu.user_id = p_user_id
        GROUP BY f.id, f.name, f.description, f.location, fu.role, f.created_at
        ORDER BY f.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: get_farm_overview
-- Lấy thông tin tổng quan của farm
-- ============================================
CREATE OR REPLACE FUNCTION get_farm_overview(p_farm_id UUID)
    RETURNS TABLE (
                      farm_id UUID,
                      farm_name VARCHAR,
                      farm_description TEXT,
                      farm_location VARCHAR,
                      total_mcus BIGINT,
                      online_mcus BIGINT,
                      offline_mcus BIGINT,
                      total_survey_points BIGINT,
                      connecting_points BIGINT,
                      connected_points BIGINT,
                      disconnected_points BIGINT,
                      created_at TIMESTAMP WITH TIME ZONE
                  ) AS $$
BEGIN
    RETURN QUERY
        SELECT
            f.id,
            f.name,
            f.description,
            f.location,
            COUNT(DISTINCT m.id) as total_mcus,
            COUNT(DISTINCT CASE WHEN m.status = 'online' THEN m.id END) as online_mcus,
            COUNT(DISTINCT CASE WHEN m.status = 'offline' THEN m.id END) as offline_mcus,
            COUNT(DISTINCT sp.id) as total_survey_points,
            COUNT(DISTINCT CASE WHEN sp.status = 'connecting' THEN sp.id END) as connecting_points,
            COUNT(DISTINCT CASE WHEN sp.status = 'connected' THEN sp.id END) as connected_points,
            COUNT(DISTINCT CASE WHEN sp.status = 'disconnected' THEN sp.id END) as disconnected_points,
            f.created_at
        FROM tbl_farms f
                 LEFT JOIN tbl_mcus m ON f.id = m.farm_id
                 LEFT JOIN tbl_survey_points sp ON m.id = sp.mcu_id
        WHERE f.id = p_farm_id
        GROUP BY f.id, f.name, f.description, f.location, f.created_at;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: get_mcu_survey_points
-- Lấy danh sách survey points của MCU
-- ============================================
CREATE OR REPLACE FUNCTION get_mcu_survey_points(p_mcu_id UUID)
    RETURNS TABLE (
                      survey_point_id UUID,
                      survey_point_name VARCHAR,
                      description TEXT,
                      status VARCHAR,
                      created_at TIMESTAMP WITH TIME ZONE,
                      updated_at TIMESTAMP WITH TIME ZONE
                  ) AS $$
BEGIN
    RETURN QUERY
        SELECT
            sp.id,
            sp.name,
            sp.description,
            sp.status,
            sp.created_at,
            sp.updated_at
        FROM tbl_survey_points sp
        WHERE sp.mcu_id = p_mcu_id
        ORDER BY sp.name;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: get_farm_mcus
-- Lấy danh sách MCUs của farm
-- ============================================
CREATE OR REPLACE FUNCTION get_farm_mcus(p_farm_id UUID)
    RETURNS TABLE (
                      mcu_id UUID,
                      mcu_code VARCHAR,
                      status VARCHAR,
                      survey_point_count BIGINT,
                      created_at TIMESTAMP WITH TIME ZONE,
                      updated_at TIMESTAMP WITH TIME ZONE
                  ) AS $$
BEGIN
    RETURN QUERY
        SELECT
            m.id,
            m.mcu_code,
            m.status,
            COUNT(sp.id) as survey_point_count,
            m.created_at,
            m.updated_at
        FROM tbl_mcus m
                 LEFT JOIN tbl_survey_points sp ON m.id = sp.mcu_id
        WHERE m.farm_id = p_farm_id
        GROUP BY m.id, m.mcu_code, m.status, m.created_at, m.updated_at
        ORDER BY m.mcu_code;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: update_mcu_status
-- Cập nhật trạng thái MCU
-- ============================================
CREATE OR REPLACE FUNCTION update_mcu_status(
    p_mcu_code VARCHAR,
    p_status VARCHAR
)
    RETURNS TABLE (
                      success BOOLEAN,
                      mcu_id UUID,
                      message TEXT
                  ) AS $$
DECLARE
    v_mcu_id UUID;
BEGIN
    -- Tìm MCU
    SELECT id INTO v_mcu_id
    FROM tbl_mcus
    WHERE mcu_code = p_mcu_code;

    IF v_mcu_id IS NULL THEN
        RETURN QUERY SELECT false, NULL::UUID, 'MCU not found'::TEXT;
        RETURN;
    END IF;

    -- Update MCU status
    UPDATE tbl_mcus
    SET
        status = p_status,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = v_mcu_id;

    RETURN QUERY SELECT true, v_mcu_id, 'MCU status updated successfully'::TEXT;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: update_survey_point_status
-- Cập nhật trạng thái survey point
-- ============================================
CREATE OR REPLACE FUNCTION update_survey_point_status(
    p_survey_point_id UUID,
    p_status VARCHAR
)
    RETURNS TABLE (
                      success BOOLEAN,
                      message TEXT
                  ) AS $$
BEGIN
    UPDATE tbl_survey_points
    SET
        status = p_status,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_survey_point_id;

    IF FOUND THEN
        RETURN QUERY SELECT true, 'Survey point status updated successfully'::TEXT;
    ELSE
        RETURN QUERY SELECT false, 'Survey point not found'::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: get_farm_structure
-- Lấy cấu trúc đầy đủ của farm (farms -> mcus -> survey_points)
-- ============================================
CREATE OR REPLACE FUNCTION get_farm_structure(p_farm_id UUID)
    RETURNS TABLE (
                      farm_id UUID,
                      farm_name VARCHAR,
                      mcu_id UUID,
                      mcu_code VARCHAR,
                      mcu_status VARCHAR,
                      survey_point_id UUID,
                      survey_point_name VARCHAR,
                      survey_point_status VARCHAR
                  ) AS $$
BEGIN
    RETURN QUERY
        SELECT
            f.id,
            f.name,
            m.id,
            m.mcu_code,
            m.status,
            sp.id,
            sp.name,
            sp.status
        FROM tbl_farms f
                 LEFT JOIN tbl_mcus m ON f.id = m.farm_id
                 LEFT JOIN tbl_survey_points sp ON m.id = sp.mcu_id
        WHERE f.id = p_farm_id
        ORDER BY m.mcu_code, sp.name;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: create_device_command
-- Tạo lệnh điều khiển thiết bị
-- ============================================
DROP FUNCTION create_device_command;
CREATE OR REPLACE FUNCTION create_device_command(
    survey_point_id UUID,
    p_device_name VARCHAR,
    p_command VARCHAR
)
    RETURNS TABLE (
                      success BOOLEAN,
                      command_id UUID,
                      message TEXT
                  ) AS $$
DECLARE
    v_command_id UUID;
BEGIN
    -- Validate command
    IF p_command NOT IN ('on', 'off') THEN
        RETURN QUERY SELECT false, NULL::UUID, 'Invalid command. Must be "on" or "off"'::TEXT;
        RETURN;
    END IF;

    -- Create command
    INSERT INTO tbl_device_commands (survey_point_id, device_name, command, status)
    VALUES (survey_point_id, p_device_name, p_command, 'pending')
    RETURNING id INTO v_command_id;

    RETURN QUERY SELECT true, v_command_id, 'Command created successfully'::TEXT;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: update_command_status
-- Cập nhật trạng thái lệnh điều khiển
-- ============================================
CREATE OR REPLACE FUNCTION update_command_status(
    p_command_id UUID,
    p_status VARCHAR
)
    RETURNS TABLE (
                      success BOOLEAN,
                      message TEXT
                  ) AS $$
BEGIN
    -- Validate status
    IF p_status NOT IN ('pending', 'sent', 'success', 'failed') THEN
        RETURN QUERY SELECT false, 'Invalid status'::TEXT;
        RETURN;
    END IF;

    UPDATE tbl_device_commands
    SET
        status = p_status,
        executed_at = CASE
                          WHEN p_status IN ('success', 'failed') THEN CURRENT_TIMESTAMP
                          ELSE executed_at
            END
    WHERE id = p_command_id;

    IF FOUND THEN
        RETURN QUERY SELECT true, 'Command status updated successfully'::TEXT;
    ELSE
        RETURN QUERY SELECT false, 'Command not found'::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: get_pending_commands
-- Lấy danh sách lệnh pending
-- ============================================
CREATE OR REPLACE FUNCTION get_pending_commands(p_limit INT DEFAULT 100)
    RETURNS TABLE (
                      command_id UUID,
                      user_id UUID,
                      username VARCHAR,
                      device_name VARCHAR,
                      command VARCHAR,
                      status VARCHAR,
                      created_at TIMESTAMP WITH TIME ZONE
                  ) AS $$
BEGIN
    RETURN QUERY
        SELECT
            dc.id,
            dc.user_id,
            u.username,
            dc.device_name,
            dc.command,
            dc.status,
            dc.created_at
        FROM tbl_device_commands dc
                 LEFT JOIN tbl_users u ON dc.user_id = u.id
        WHERE dc.status = 'pending'
        ORDER BY dc.created_at ASC
        LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: get_command_history
-- Lấy lịch sử lệnh điều khiển
-- ============================================
DROP FUNCTION get_command_history;
CREATE OR REPLACE FUNCTION get_command_history(
    survey_point_id UUID DEFAULT NULL,
    p_device_name VARCHAR DEFAULT NULL,
    p_limit INT DEFAULT 50
)
    RETURNS TABLE (
                      command_id UUID,
                      survey_point_id UUID,
                      username VARCHAR,
                      device_name VARCHAR,
                      command VARCHAR,
                      status VARCHAR,
                      executed_at TIMESTAMP WITH TIME ZONE,
                      created_at TIMESTAMP WITH TIME ZONE
                  ) AS $$
BEGIN
    RETURN QUERY
        SELECT
            dc.id,
            dc.survey_point_id,
            u.username,
            dc.device_name,
            dc.command,
            dc.status,
            dc.executed_at,
            dc.created_at
        FROM tbl_device_commands dc
                 LEFT JOIN tbl_survey_points u ON dc.user_id = u.id
        WHERE
            (survey_point_id IS NULL OR dc.user_id = survey_point_id)
          AND (p_device_name IS NULL OR dc.device_name = p_device_name)
        ORDER BY dc.created_at DESC
        LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: check_user_farm_permission
-- Kiểm tra quyền của user với farm
-- ============================================
CREATE OR REPLACE FUNCTION check_user_farm_permission(
    p_user_id UUID,
    p_farm_id UUID,
    p_required_role VARCHAR DEFAULT 'viewer'
)
    RETURNS BOOLEAN AS $$
DECLARE
    v_user_role VARCHAR;
    v_role_level INT;
    v_required_level INT;
BEGIN
    -- Get user role
    SELECT role INTO v_user_role
    FROM tbl_farm_users
    WHERE user_id = p_user_id AND farm_id = p_farm_id;

    IF v_user_role IS NULL THEN
        RETURN false;
    END IF;

    -- Convert roles to levels: viewer=1, manager=2, owner=3
    v_role_level := CASE v_user_role
                        WHEN 'owner' THEN 3
                        WHEN 'manager' THEN 2
                        WHEN 'viewer' THEN 1
                        ELSE 0
        END;

    v_required_level := CASE p_required_role
                            WHEN 'owner' THEN 3
                            WHEN 'manager' THEN 2
                            WHEN 'viewer' THEN 1
                            ELSE 0
        END;

    RETURN v_role_level >= v_required_level;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: get_user_by_username
-- Lấy thông tin user theo username (for authentication)
-- ============================================
CREATE OR REPLACE FUNCTION get_user_by_username(p_username VARCHAR)
    RETURNS TABLE (
                      user_id UUID,
                      username VARCHAR,
                      email VARCHAR,
                      phone VARCHAR,
                      password_hash VARCHAR,
                      full_name VARCHAR,
                      created_at TIMESTAMP WITH TIME ZONE
                  ) AS $$
BEGIN
    RETURN QUERY
        SELECT
            u.id,
            u.username,
            u.email,
            u.phone,
            u.password_hash,
            u.full_name,
            u.created_at
        FROM tbl_users u
        WHERE u.username = p_username;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: get_user_by_email
-- Lấy thông tin user theo email
-- ============================================
CREATE OR REPLACE FUNCTION get_user_by_email(p_email VARCHAR)
    RETURNS TABLE (
                      user_id UUID,
                      username VARCHAR,
                      email VARCHAR,
                      phone VARCHAR,
                      password_hash VARCHAR,
                      full_name VARCHAR,
                      created_at TIMESTAMP WITH TIME ZONE
                  ) AS $$
BEGIN
    RETURN QUERY
        SELECT
            u.id,
            u.username,
            u.email,
            u.phone,
            u.password_hash,
            u.full_name,
            u.created_at
        FROM tbl_users u
        WHERE u.email = p_email;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: create_farm_with_owner
-- Tạo farm mới và gán owner
-- ============================================
CREATE OR REPLACE FUNCTION create_farm_with_owner(
    p_user_id UUID,
    p_farm_name VARCHAR,
    p_description TEXT DEFAULT NULL,
    p_location VARCHAR DEFAULT NULL
)
    RETURNS TABLE (
                      success BOOLEAN,
                      farm_id UUID,
                      message TEXT
                  ) AS $$
DECLARE
    v_farm_id UUID;
BEGIN
    -- Create farm
    INSERT INTO tbl_farms (name, description, location)
    VALUES (p_farm_name, p_description, p_location)
    RETURNING id INTO v_farm_id;

    -- Add user as owner
    INSERT INTO tbl_farm_users (user_id, farm_id, role)
    VALUES (p_user_id, v_farm_id, 'owner');

    RETURN QUERY SELECT true, v_farm_id, 'Farm created successfully'::TEXT;
EXCEPTION
    WHEN OTHERS THEN
        RETURN QUERY SELECT false, NULL::UUID, SQLERRM::TEXT;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: add_user_to_farm
-- Thêm user vào farm với role
-- ============================================
CREATE OR REPLACE FUNCTION add_user_to_farm(
    p_user_id UUID,
    p_farm_id UUID,
    p_role VARCHAR DEFAULT 'viewer'
)
    RETURNS TABLE (
                      success BOOLEAN,
                      message TEXT
                  ) AS $$
BEGIN
    -- Validate role
    IF p_role NOT IN ('owner', 'manager', 'viewer') THEN
        RETURN QUERY SELECT false, 'Invalid role. Must be owner, manager, or viewer'::TEXT;
        RETURN;
    END IF;

    -- Add user to farm
    INSERT INTO tbl_farm_users (user_id, farm_id, role)
    VALUES (p_user_id, p_farm_id, p_role)
    ON CONFLICT (user_id, farm_id)
        DO UPDATE SET role = p_role;

    RETURN QUERY SELECT true, 'User added to farm successfully'::TEXT;
EXCEPTION
    WHEN foreign_key_violation THEN
        RETURN QUERY SELECT false, 'User or farm not found'::TEXT;
    WHEN OTHERS THEN
        RETURN QUERY SELECT false, SQLERRM::TEXT;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: remove_user_from_farm
-- Xóa user khỏi farm
-- ============================================
CREATE OR REPLACE FUNCTION remove_user_from_farm(
    p_user_id UUID,
    p_farm_id UUID
)
    RETURNS TABLE (
                      success BOOLEAN,
                      message TEXT
                  ) AS $$
BEGIN
    DELETE FROM tbl_farm_users
    WHERE user_id = p_user_id AND farm_id = p_farm_id;

    IF FOUND THEN
        RETURN QUERY SELECT true, 'User removed from farm successfully'::TEXT;
    ELSE
        RETURN QUERY SELECT false, 'User-farm relationship not found'::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: get_mcu_by_code
-- Lấy thông tin MCU theo code
-- ============================================
CREATE OR REPLACE FUNCTION get_mcu_by_code(p_mcu_code VARCHAR)
    RETURNS TABLE (
                      mcu_id UUID,
                      mcu_code VARCHAR,
                      farm_id UUID,
                      farm_name VARCHAR,
                      status VARCHAR,
                      survey_point_count BIGINT,
                      created_at TIMESTAMP WITH TIME ZONE,
                      updated_at TIMESTAMP WITH TIME ZONE
                  ) AS $$
BEGIN
    RETURN QUERY
        SELECT
            m.id,
            m.mcu_code,
            f.id,
            f.name,
            m.status,
            COUNT(sp.id) as survey_point_count,
            m.created_at,
            m.updated_at
        FROM tbl_mcus m
                 INNER JOIN tbl_farms f ON m.farm_id = f.id
                 LEFT JOIN tbl_survey_points sp ON m.id = sp.mcu_id
        WHERE m.mcu_code = p_mcu_code
        GROUP BY m.id, m.mcu_code, f.id, f.name, m.status, m.created_at, m.updated_at;
END;
$$ LANGUAGE plpgsql;
