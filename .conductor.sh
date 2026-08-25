# 1. تغییر نام برنچ محلی از mine به main
git branch -m mine main

# 2. دریافت آخرین تغییرات از مخزن راه‌دور
git fetch origin

# 3. تنظیم برنچ main برای دنبال کردن origin/main
git branch -u origin/main main

# 4. تنظیم HEAD راه‌دور به برنچ main
git remote set-head origin main