/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../screens/export.dart';
import '../utils/export.dart';
import 'package:ywt_private/ywt_private.dart' as ywt;

import 'package:open_ui/open_ui.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class SettingsFAB extends FloatingActionButton {
  final EzCP config;

  SettingsFAB(this.config, {required BuildContext parentContext, super.key})
      : super(
          child: EzIcon(config, Icons.settings),
          onPressed: () => parentContext.goNamed(settingsHubPath),
          tooltip: l10n(config).gSettingsHint,
        );
}

EzUpdaterFAB updater(EzCP config) => EzUpdaterFAB(
      config,
      appVersion: '1.0.1',
      versionSource:
          'https://raw.githubusercontent.com/YWT-LLC/wyrd_web/refs/heads/main/APP_VERSION',
      gPlay: ywt.wyrdWebGPlay,
      appStore: ywt.wyrdWebAppStore,
      github: ywt.wyrdWebReleases,
    );
