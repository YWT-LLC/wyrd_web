import 'lang.dart';

// ignore_for_file: type=lint

/// The translations for Spanish Castilian (`es`).
class LangEs extends Lang {
  LangEs([String locale = 'es']) : super(locale);

  @override
  String get gLogoHint =>
      'Logotipo de Empathetic LLC, un reloj de arena en 2D. Actívalo para ir a la página principal';

  @override
  String get gSettingsHint => 'Abrir la página de configuración';
}
