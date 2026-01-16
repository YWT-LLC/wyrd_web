/* wyrd_web
 * Copyright (c) 2026 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../utils/export.dart';

import 'package:flutter/material.dart';
import 'package:empathetech_flutter_ui/empathetech_flutter_ui.dart';

class WyrdWebScaffold extends StatelessWidget {
  final Widget body;

  /// [FloatingActionButton]
  final Widget? fab;

  /// Standardized [Scaffold] for all of the EFUI example app's screens
  const WyrdWebScaffold({
    super.key,
    required this.body,
    this.fab,
  });

  @override
  Widget build(BuildContext context) {
    // Gather the theme data //

    final bool isLefty = EzConfig.get(isLeftyKey) ?? false;

    final Size appBarTextSize = ezTextSize(
      appName,
      style: Theme.of(context).appBarTheme.titleTextStyle,
      context: context,
    );

    final double toolbarHeight = appBarTextSize.height + EzConfig.marginVal;

    // Return the build //

    return EzAdaptiveParent(
      small: SelectionArea(
        child: Scaffold(
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
              actions: isLefty ? const <Widget>[EzBackAction()] : null,
            ),
          ),

          // Body
          body: body,

          // FAB
          floatingActionButton: fab,
          floatingActionButtonLocation: isLefty
              ? FloatingActionButtonLocation.startFloat
              : FloatingActionButtonLocation.endFloat,

          // Prevents the keyboard from pushing the body up
          resizeToAvoidBottomInset: false,
        ),
      ),
    );
  }
}
