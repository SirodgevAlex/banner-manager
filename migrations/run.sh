#!bin/sh

psql -U postgres -d banner_manager_tables -f 001_init_users.sql
psql -U postgres -d banner_manager_tables -f 002_init_banners.sql
psql -U postgres -d banner_manager_tables -f 003_get_users.sql
psql -U postgres -d banner_manager_tables -f 004_get_banners.sql