#!/usr/bin/env bash
DB="/home/ctn/.config/opencode/custom/memory/memory.db"

printf "\n-=-=-=--=-=-=--=-=-=-\n\n"

echo "=== HIPPOCAMPUS ==="
echo "--- TOP 5 BY SCORE ---"
sqlite3 -header "$DB" \
  "SELECT id, substr(content,1,100) AS content, printf('%.2f',score) AS score, created_at FROM hippocampus ORDER BY score DESC LIMIT 5;"
echo "--- NEWEST 5 ---"
sqlite3 -header "$DB" \
  "SELECT id, substr(content,1,100) AS content, printf('%.2f',score) AS score, created_at FROM hippocampus ORDER BY created_at DESC LIMIT 5;"

echo ""
echo "=== NEOCORTEX ==="
echo "--- TOP 5 BY SCORE ---"
sqlite3 -header "$DB" \
  "SELECT id, substr(content,1,100) AS content, printf('%.2f',score) AS score, created_at FROM neocortex ORDER BY score DESC LIMIT 5;"
echo "--- NEWEST 5 ---"
sqlite3 -header "$DB" \
  "SELECT id, substr(content,1,100) AS content, printf('%.2f',score) AS score, created_at FROM neocortex ORDER BY created_at DESC LIMIT 5;"
