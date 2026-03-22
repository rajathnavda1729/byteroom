CREATE TYPE member_role AS ENUM ('admin', 'member');

ALTER TABLE chat_members
    ADD COLUMN role member_role NOT NULL DEFAULT 'member';
