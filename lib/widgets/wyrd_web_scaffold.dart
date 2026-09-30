/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import 'dart:math';
import './export.dart';

import 'package:open_ui/open_ui.dart';
import 'package:flutter/material.dart';

class WyrdWebScaffold extends StatelessWidget {
  final EzCP config;
  final Widget body;
  final List<HybridAction> actions;
  final List<Widget>? settingsFABs;
  final bool isHome;

  const WyrdWebScaffold(
    this.config, {
    super.key,
    required this.body,
    required this.actions,
    this.settingsFABs,
    this.isHome = false,
  });

  @override
  Widget build(BuildContext context) {
    // Define the build data //

    final double toolbarHeight = max(
            config.iconSize,
            ezTextSize(
              config,
              text: config.ezL10n.gSettings,
              style: config.bodyStyle,
              textScaler: MediaQuery.textScalerOf(context),
            ).height) +
        config.padding;

    // Define custom functions //

    Iterable<Widget> fabActions() => actions.map((HybridAction action) {
          final Widget core = Padding(
            padding: EdgeInsets.only(top: config.spacing),
            child: FloatingActionButton(
              heroTag: '${action.label}_FAB',
              onPressed: action.onPressed,
              tooltip: action.label,
              child: EzIcon(config, action.icon),
            ),
          );

          return action.menuController == null
              ? core
              : MenuAnchor(
                  controller: action.menuController!,
                  menuChildren: action.menuChildren!,
                  child: core,
                );
        });

    List<Widget> toolbarActions() => actions.map((HybridAction action) {
          final Widget core = Padding(
              padding: EdgeInsets.symmetric(horizontal: config.spacing / 2),
              child: EzTextIconButton(
                config,
                label: action.label,
                icon: EzIcon(config, action.icon),
                onPressed: action.onPressed,
              ));

          return action.menuController == null
              ? core
              : MenuAnchor(
                  controller: action.menuController!,
                  menuChildren: action.menuChildren!,
                  child: core,
                );
        }).toList();

    // Return the build //

    return EzAdaptiveParent(
      small: EzScaffold(
        config,
        body: body,
        fabs: <Widget>[
          updater(config),
          ...fabActions(),
          if (settingsFABs != null) ...settingsFABs!,
          ...config.backFABs(isHome),
        ],
      ),
      medium: EzScaffold(
        config,
        appBar: PreferredSize(
          preferredSize: Size(double.infinity, toolbarHeight),
          child: EzAppBar(
            config,
            height: toolbarHeight,
            title: EzScrollView(
              config,
              reverseHands: true,
              showScrollHint: true,
              thumbVisibility: false,
              scrollDirection: Axis.horizontal,
              mainAxisAlignment: MainAxisAlignment.center,
              children: toolbarActions(),
            ),
          ),
        ),
        body: body,
        fabs: <Widget>[
          updater(config),
          if (settingsFABs != null) ...settingsFABs!,
          ...config.backFABs(isHome),
        ],
      ),
    );
  }
}
