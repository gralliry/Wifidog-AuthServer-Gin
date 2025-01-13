/*
 Navicat Premium Dump SQL

 Source Server         : Wifidog
 Source Server Type    : SQLite
 Source Server Version : 3045000 (3.45.0)
 Source Schema         : main

 Target Server Type    : SQLite
 Target Server Version : 3045000 (3.45.0)
 File Encoding         : 65001

 Date: 13/01/2025 23:25:49
*/

PRAGMA foreign_keys = false;

-- ----------------------------
-- Table structure for conn
-- ----------------------------
DROP TABLE IF EXISTS "conn";
CREATE TABLE "conn" (
  "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  "token" TEXT NOT NULL,
  "user_id" text NOT NULL,
  "net_id" text NOT NULL DEFAULT 0,
  "ip" TEXT NOT NULL,
  "mac" TEXT NOT NULL,
  "incoming" INTEGER NOT NULL DEFAULT 0,
  "outgoing" INTEGER NOT NULL DEFAULT 0,
  "start_time" INTEGER NOT NULL DEFAULT (0),
  "end_time" INTEGER NOT NULL DEFAULT (0),
  "is_expire" integer NOT NULL DEFAULT (0),
  FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON DELETE SET NULL ON UPDATE CASCADE,
  FOREIGN KEY ("net_id") REFERENCES "net" ("id") ON DELETE SET NULL ON UPDATE CASCADE,
  UNIQUE ("id" ASC),
  UNIQUE ("token" ASC)
);

-- ----------------------------
-- Table structure for net
-- ----------------------------
DROP TABLE IF EXISTS "net";
CREATE TABLE "net" (
  "id" INTEGER NOT NULL,
  "sid" TEXT NOT NULL,
  "address" TEXT NOT NULL,
  "port" INTEGER NOT NULL,
  "sys_uptime" INTEGER NOT NULL DEFAULT (0),
  "sys_memfree" INTEGER NOT NULL DEFAULT (0),
  "sys_load" REAL NOT NULL DEFAULT (0.0),
  "wifidog_uptime" INTEGER NOT NULL DEFAULT (0),
  PRIMARY KEY ("id", "sid"),
  UNIQUE ("id" ASC)
);

-- ----------------------------
-- Table structure for sqlite_sequence
-- ----------------------------
DROP TABLE IF EXISTS "sqlite_sequence";
CREATE TABLE "sqlite_sequence" (
  "name",
  "seq"
);

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS "user";
CREATE TABLE "user" (
  "id" integer NOT NULL DEFAULT 0,
  "account" TEXT NOT NULL,
  "password" TEXT NOT NULL,
  PRIMARY KEY ("id", "account"),
  UNIQUE ("id" ASC),
  UNIQUE ("account" ASC)
);

-- ----------------------------
-- Auto increment value for conn
-- ----------------------------
UPDATE "sqlite_sequence" SET seq = 18 WHERE name = 'conn';

PRAGMA foreign_keys = true;
