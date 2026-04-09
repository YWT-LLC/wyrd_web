/* wyrd_web
 * Copyright (c) 2026 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import './export.dart';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:empathetech_flutter_ui/empathetech_flutter_ui.dart';

class WyrdWebScaffold extends StatelessWidget {
  /// [Scaffold.body] passthrough
  final Widget body;

  /// [FloatingActionButton]s to add on top of the [EzUpdaterFAB]
  /// BYO spacing widgets
  final List<Widget>? fabs;

  /// Standardized [Scaffold] for all of the EFUI example app's screens
  const WyrdWebScaffold(this.body, {super.key, this.fabs});

  @override
  Widget build(BuildContext context) {
    // Return the build //

    return EzAdaptiveParent(
      small: Consumer<EzConfigProvider>(
        builder: (_, EzConfigProvider config, __) => SelectionArea(
          child: Scaffold(
            key: ValueKey<int>(config.seed),
            body: body,
            floatingActionButton: Column(
              mainAxisSize: MainAxisSize.min,
              children: <Widget>[
                updater,
                if (fabs != null) ...fabs!,
                if (config.layout.showBackFAB &&
                    ezRootNav.currentState!.canPop()) ...<Widget>[
                  config.layout.spacer,
                  const EzBackFAB(),
                ],
              ],
            ),
            floatingActionButtonLocation: EzConfig.isLefty
                ? FloatingActionButtonLocation.startFloat
                : FloatingActionButtonLocation.endFloat,
            resizeToAvoidBottomInset: false,
          ),
        ),
      ),
    );
  }
}
