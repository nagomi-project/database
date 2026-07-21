-- name: IsModuleEnabled :one
-- Checks if a specific module is enabled.
SELECT COALESCE((
    SELECT
        enabled
    FROM guild_modules
    WHERE
        guild_id = @guild_id
        AND module_type = @module_type
), FALSE)::BOOLEAN AS enabled;

-- name: ToggleModule :exec
-- Toggles if a module is enabled.
-- Modules that are not in the database are considered disabled by default.
INSERT INTO guild_modules (
    guild_id,
    module_type,
    enabled
)
VALUES (
    @guild_id,
    @module_type,
    @enabled
)
ON CONFLICT (guild_id, module_type) DO UPDATE SET
    updated_at = now(),
    enabled = EXCLUDED.enabled;