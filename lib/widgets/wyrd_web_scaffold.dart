/* wyrd_web
 * Copyright (c) 2026 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../utils/export.dart';
import './export.dart';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:empathetech_flutter_ui/empathetech_flutter_ui.dart';

class WyrdWebScaffold extends StatelessWidget {
  final Widget body;

  /// [FloatingActionButton]s to add on top of the [EzUpdaterFAB]
  /// BYO spacing widgets
  final List<Widget>? fabs;

  /// Standardized [Scaffold] for all of the EFUI example app's screens
  const WyrdWebScaffold(this.body, {super.key, this.fabs});

  @override
  Widget build(BuildContext context) {
    // Gather the contextual theme data //

    final Size appBarTextSize = ezTextSize(
      appName,
      style: Theme.of(context).appBarTheme.titleTextStyle,
      context: context,
    );

    final double toolbarHeight = appBarTextSize.height + EzConfig.marginVal;

    // Return the build //

    return EzAdaptiveParent(
      small: Consumer<EzConfigProvider>(
        builder: (_, EzConfigProvider provider, __) => SelectionArea(
          child: Scaffold(
            key: ValueKey<int>(provider.seed),

            // AppBar
            appBar: PreferredSize(
              preferredSize: Size(double.infinity, toolbarHeight),
              child: AppBar(
                excludeHeaderSemantics: true,
                toolbarHeight: toolbarHeight,

                // Leading (aka left)
                leading: null,
                leadingWidth: toolbarHeight,

                // Title
                title: const Text(appName),

                // Actions (aka trailing aka right)
                actions:
                    EzConfig.isLefty ? const <Widget>[EzBackAction()] : null,
              ),
            ),

            // Body
            body: body,

            // FAB
            floatingActionButton: Column(
              mainAxisSize: MainAxisSize.min,
              children: <Widget>[updater, if (fabs != null) ...fabs!],
            ),
            floatingActionButtonLocation: EzConfig.isLefty
                ? FloatingActionButtonLocation.startFloat
                : FloatingActionButtonLocation.endFloat,

            // Prevents the keyboard from pushing the body up
            resizeToAvoidBottomInset: false,
          ),
        ),
      ),
    );
  }
}
