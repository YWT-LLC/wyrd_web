/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import 'package:open_ui/open_ui.dart';
import 'package:flutter/material.dart';

enum Services { smokeSignal }

extension ServiceConfig on Services {
  String get name => switch (this) {
        Services.smokeSignal => 'Smoke Signal',
      };

  Widget icon(EzCP config) => switch (this) {
        Services.smokeSignal => EzIcon(config, Icons.fireplace),
      };

  void Function() get startup => switch (this) {
        Services.smokeSignal => doNothing,
      };
}
