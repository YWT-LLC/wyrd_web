/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../screens/export.dart';
import '../utils/export.dart';
import '../widgets/export.dart';
import 'package:ywt_private/ywt_private.dart' as ywt;

import 'package:open_ui/open_ui.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  // Define the build data //

  ValueNotifier<String> readout = ValueNotifier<String>('');
  bool showReadout = true;

  // Set the page title //

  @override
  void initState() {
    super.initState();
    ezWindowNamer(appName);
  }

  // Return the build //

  @override
  Widget build(BuildContext context) {
    return Consumer<EzCP>(
      builder: (_, EzCP config, __) => WyrdWebScaffold(
        config,
        body: EzScreen(
          config,
          alignment: Alignment.center,
          child: EzScrollView(
            config,
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              // Options
              EzWrap(
                children: Services.values
                    .map((Services service) => EzElevatedIconButton(
                          config,
                          label: service.name,
                          icon: service.icon(config),
                          onPressed: () => context.goNamed(runningPath, extra: service),
                        ))
                    .toList(),
              ),

              // Footer
              EzFooter(config, a11howPath: ywt.wyrdWebContributeA11),
            ],
          ),
        ),
        isHome: true,
        actions: <HybridAction>[
          HybridAction(
            label: config.ezL10n.gSettings,
            icon: Icons.settings,
            onPressed: () => context.goNamed(settingsHubPath),
          ),
        ],
      ),
    );
  }
}
