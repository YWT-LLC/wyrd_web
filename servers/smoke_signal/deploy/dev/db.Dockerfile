# wyrd_web
# Copyright (c) 2025 Empathetech LLC. All rights reserved.
# See LICENSE for distribution and usage details.

FROM postgres

COPY fixtures/db_setup.sql /docker-entrypoint-initdb.d/