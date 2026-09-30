/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import 'screens/export.dart';
import 'utils/export.dart';

import 'package:open_ui/open_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() async {
  // Configure the app //

  WidgetsFlutterBinding.ensureInitialized();
  await SystemChrome.setPreferredOrientations(DeviceOrientation.values);

  // Initialize EzConfig //

  EzCM.init(
    appName: appName,
    androidPackage: androidPackage,
    assetPaths: <String>{},
    orientations: DeviceOrientation.values,
    localeFallback: americanEnglish,
    l10nFallback: await OUILang.delegate.load(americanEnglish),
    preferences: await SharedPreferencesWithCache.create(
      cacheOptions: SharedPreferencesWithCacheOptions(
        allowList: allEZConfigKeys.keys.toSet(), // TODO
      ),
    ),
    defaults: isMobile() ? ywtMobileConfig : ywtDesktopConfig,
  );

  // Run the app //

  final (Locale storedLocale, OUILang storedOUILang) = await ezStoredL10n();

  runApp(WyrdWeb(
    storedLocale,
    storedOUILang,
    await Lang.delegate.load(storedLocale),
  ));
}

class WyrdWeb extends StatelessWidget {
  final Locale storedLocale;
  final OUILang storedOUILang;
  final Lang storedLang;

  const WyrdWeb(this.storedLocale, this.storedOUILang, this.storedLang, {super.key});

  @override
  Widget build(BuildContext context) => EzConfigurableApp(
        localizationsDelegates: ezLocalizationsDelegates(Lang.localizationsDelegates),
        supportedLocales: Lang.supportedLocales,
        locale: storedLocale,
        el10n: storedOUILang,
        appCache: WyrdWebCache(storedLocale, storedLang),
        routerConfig: GoRouter(
          initialLocation: homePath,
          errorBuilder: (_, __) => const ErrorScreen(),
          routes: <RouteBase>[
            // Home/select service
            GoRoute(
              path: homePath,
              name: homePath,
              pageBuilder: (BuildContext pbc, GoRouterState pbs) => ezPageBuilder(
                configWatcher(pbc),
                pbc,
                pbs,
                const HomeScreen(),
              ),
              routes: <RouteBase>[
                // Run service
                GoRoute(
                  path: runningPath,
                  name: runningPath,
                  pageBuilder: (BuildContext pbc, GoRouterState pbs) => ezPageBuilder(
                    configWatcher(pbc),
                    pbc,
                    pbs,
                    RunningScreen(pbs.extra as Services),
                  ),
                ),

                // Settings
                GoRoute(
                  path: settingsHubPath,
                  name: settingsHubPath,
                  pageBuilder: (BuildContext pbc, GoRouterState pbs) => ezPageBuilder(
                    configWatcher(pbc),
                    pbc,
                    pbs,
                    const SettingsHubScreen(),
                  ),
                ),
              ],
            ),
          ],
        ),
      );
}
