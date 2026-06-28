-- migrations/000009_create_quiz_leaderboard.down.sql

DROP TABLE IF EXISTS quiz_leaderboard;

ALTER TABLE quiz_attempts DROP COLUMN IF EXISTS percentage;
