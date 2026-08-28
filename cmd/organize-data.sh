#!/bin/bash
echo "📁 Organizing data files..."

# ایجاد پوشه‌های لازم
mkdir -p data
mkdir -p data/json
mkdir -p data/yaml

# انتقال فایل‌های داده به پوشه‌های مربوطه
mv *.json data/json/ 2>/dev/null || true
mv *.yaml data/yaml/ 2>/dev/null || true
mv *.yml data/yaml/ 2>/dev/null || true

# اطمینان از وجود .env و liara.json در ریشه
touch .env
touch liara.json

echo "✅ Data files organized."
