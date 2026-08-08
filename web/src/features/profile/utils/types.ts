export type ActivityLevel = 'sedentary' | 'light' | 'moderate' | 'active' | 'very_active';

export type Sex = 'male' | 'female';

export type DailyTargets = {
  calories: number;
  protein: number;
  fat: number;
  carbs: number;
  bmr: number;
  tdee: number;
};

export type UserProfile = {
  weight_kg: number | null;
  height_cm: number | null;
  age: number | null;
  sex: Sex | '';
  activity_level: ActivityLevel | '';
  daily_targets: DailyTargets | null;
};

export type ProfileInput = {
  weight_kg: number;
  height_cm: number;
  age: number;
  sex: Sex;
  activity_level: ActivityLevel;
};

export const ACTIVITY_OPTIONS: { value: ActivityLevel; label: string }[] = [
  { value: 'sedentary', label: 'Сидячий спосіб життя' },
  { value: 'light', label: 'Легка активність (1–3 р/тиж)' },
  { value: 'moderate', label: 'Помірна активність (3–5 р/тиж)' },
  { value: 'active', label: 'Висока активність (6–7 р/тиж)' },
  { value: 'very_active', label: 'Дуже висока / фізична робота' },
];

export const SEX_OPTIONS: { value: Sex; label: string }[] = [
  { value: 'male', label: 'Чоловік' },
  { value: 'female', label: 'Жінка' },
];
