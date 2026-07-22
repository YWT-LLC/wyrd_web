# wyrd_web
# Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
# See LICENSE for distribution and usage details.

FROM postgres

COPY fixtures/db_setup.sql /docker-entrypoint-initdb.d/