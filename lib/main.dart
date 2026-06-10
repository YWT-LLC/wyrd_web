/* wyrd_web
 * Copyright (c) 2026 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

import 'screens/export.dart';
import 'utils/export.dart';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:empathetech_flutter_ui/empathetech_flutter_ui.dart';
import 'package:flutter_localized_locales/flutter_localized_locales.dart';

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
    l10nFallback: await EFUILang.delegate.load(americanEnglish),
    preferences: await SharedPreferencesWithCache.create(
      cacheOptions: SharedPreferencesWithCacheOptions(
        allowList: allEZConfigKeys.keys.toSet(),
      ),
    ),
    defaults: isMobile() ? empathMobileConfig : empathDesktopConfig,
  );

  // Run the app //

  final (Locale storedLocale, EFUILang storedEFUILang) = await ezStoredL10n();

  runApp(WyrdWeb(
    storedLocale,
    storedEFUILang,
    await Lang.delegate.load(storedLocale),
  ));
}

class WyrdWeb extends StatelessWidget {
  final Locale storedLocale;
  final EFUILang storedEFUILang;
  final Lang storedLang;

  const WyrdWeb(
    this.storedLocale,
    this.storedEFUILang,
    this.storedLang, {
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    return EzConfigurableApp(
      localizationsDelegates: <LocalizationsDelegate<dynamic>>{
        const LocaleNamesLocalizationsDelegate(),
        ...EFUILang.localizationsDelegates,
        ...Lang.localizationsDelegates,
      },
      supportedLocales: Lang.supportedLocales,
      locale: storedLocale,
      el10n: storedEFUILang,
      appCache: WyrdWebCache(storedLocale, storedLang),
      routerConfig: GoRouter(
        initialLocation: homePath,
        errorBuilder: (_, GoRouterState state) => const ErrorScreen(),
        routes: <RouteBase>[
          // Home
          GoRoute(
            path: homePath,
            name: homePath,
            pageBuilder: (BuildContext pbc, GoRouterState pbs) =>
                ezPageBuilder(configWatcher(pbc), pbc, pbs, const HomeScreen()),
            routes: <RouteBase>[
              // Settings
              GoRoute(
                path: settingsHubPath,
                name: settingsHubPath,
                pageBuilder: (BuildContext pbc, GoRouterState pbs) =>
                    ezPageBuilder(configWatcher(pbc), pbc, pbs, const SettingsHubScreen()),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
