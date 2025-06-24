/* wyrd_web
 * Copyright (c) 2025 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import '../utils/export.dart';
import '../widgets/export.dart';

import 'package:flutter/material.dart';
import 'package:empathetech_flutter_ui/empathetech_flutter_ui.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  // Gather theme data //

  final double margin = EzConfig.get(marginKey);
  final double spacing = EzConfig.get(spacingKey);

  late final TextTheme textTheme = Theme.of(context).textTheme;

  // Define the build data //

  ValueNotifier<String> readout = ValueNotifier<String>('');
  bool showReadout = true;

  // Set the page title //

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    ezWindowNamer(context, appTitle);
  }

  // Return the build //

  @override
  Widget build(BuildContext context) {
    return WyrdWebScaffold(
      body: EzScreen(
        child: EzScrollView(children: <Widget>[
          if (spacing > margin) EzSpacer(space: spacing - margin),

          // Launch //
          // Smoke Signal
          EzElevatedIconButton(
            onPressed: () async {
              await ezCmd(
                'kubectl config use-context docker-desktop -n default',
                dir:
                    '/Users/mwaldron/repos/flutter/wyrd_web/servers/smoke_signal',
                onSuccess: doNothing,
                onFailure: (_) {
                  return;
                },
                readout: readout,
              ); // Needs fix: don't hard-code dir

              await ezCmd(
                'skaffold dev',
                dir:
                    '/Users/mwaldron/repos/flutter/wyrd_web/servers/smoke_signal',
                onSuccess: doNothing,
                onFailure: (_) {
                  return;
                },
                readout: readout,
              ); // Ditto
            },
            icon: const Icon(Icons.launch),
            label: 'Smoke Signal',
          ),
          const EzDivider(),

          // Monitor //

          // CLI
          EzRow(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              EzText(
                'Console',
                style: textTheme.titleLarge,
                textAlign: TextAlign.center,
              ),
              EzMargin(vertical: false),
              EzIconButton(
                onPressed: () => setState(() => showReadout = !showReadout),
                icon: EzIcon(
                  showReadout ? Icons.arrow_drop_up : Icons.arrow_drop_down,
                ),
              ),
            ],
          ),
          EzMargin(),

          // Readout
          Visibility(
            visible: showReadout,
            child: Container(
              constraints: BoxConstraints(
                minWidth: widthOf(context) * 0.667,
                maxWidth: widthOf(context) * 0.667,
                maxHeight: heightOf(context) / 2,
              ),
              padding: EdgeInsets.all(EzConfig.get(marginKey)),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                borderRadius: ezRoundEdge,
              ),
              child: ValueListenableBuilder<String>(
                valueListenable: readout,
                builder: (_, String value, __) => EzScrollView(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  child: Text(
                    value,
                    style: textTheme.bodyLarge,
                    textAlign: TextAlign.start,
                  ),
                ),
              ),
            ),
          ),
          const EzSeparator(),
        ]),
      ),
      fab: SettingsFAB(context),
    );
  }
}
