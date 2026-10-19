# Aupa — Antigravity Looping Prompt (বাংলা নির্দেশনা)

এই ফাইলের ইংরেজি prompt-টি Antigravity-তে paste করো। English prompt রাখা হয়েছে যাতে coding agent-এর technical instruction স্পষ্ট থাকে।

## কীভাবে ব্যবহার করবে
1. Antigravity-তে **Aupa project root folder** খুলবে।
2. `AGENTS.md` root-এ রাখবে; এই prompt-টি `docs/`-এ রাখবে।
3. `docs/IMPLEMENTATION_STATUS_TEMPLATE.md` থেকে `docs/IMPLEMENTATION_STATUS.md` বানাবে—অথবা agent-কে বানাতে দেবে।
4. `AUPA_LOOPING_ANTIGRAVITY_PROMPT.md` খুলে `COPY FROM HERE` থেকে `COPY ENDS HERE` পর্যন্ত কপি করে Antigravity-তে paste করবে।
5. আগে থেকে তৈরি Flutter UI মুছে ফেলতে বা একবারে পুরো app rewrite করতে দেবে না।

## কাজের বাধ্যতামূলক ক্রম
1. **Audit:** আগে বর্তমান project বুঝবে, existing UI সংরক্ষণ করবে এবং task list বানাবে।
2. **Frontend:** Flutter/Dart screen, navigation, loading/error state, repository interface ও mock-data flow সম্পূর্ণ করবে।
3. **Backend:** Go API, custom authentication, role/ownership authorization, order rules ও tests তৈরি করবে। Rust শুধু বাস্তবসম্মত dispatch/optimization কাজ থাকলে করবে।
4. **Supabase:** PostgreSQL migrations/schema/indexes/policies/seed data এবং Go-এর real database adapter যুক্ত করবে। Supabase Auth ব্যবহার করবে না।
5. **Integration & QA:** Flutter ↔ Go ↔ Supabase পুরো flow, security negative tests ও build যাচাই করবে।

## Loop-এর নিয়ম
প্রতি iteration-এ একটিমাত্র ছোট task নেবে → implement করবে → test চালাবে → বাস্তব ফল লিখবে → `docs/IMPLEMENTATION_STATUS.md` update করবে → তারপর পরের task নেবে। আগের phase অসম্পূর্ণ রেখে পরের phase-এ যাবে না। Test না চালিয়ে pass বলবে না। Blocker থাকলে লিখে রাখবে এবং একই phase-এর independent task চালিয়ে যাবে। সব task শেষ হলে বা নিরাপদে এগোনো অসম্ভব হলে থামবে; অনন্তকাল একই কাজ ঘুরিয়ে করবে না।

**মূল prompt:** `docs/AUPA_LOOPING_ANTIGRAVITY_PROMPT.md`
