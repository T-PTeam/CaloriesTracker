import { useEffect, useState, type FormEvent } from 'react';
import { useI18n } from '../../../i18n/I18nProvider';
import {
  ActionRow,
  FormField,
  FormInput,
  FormPanel,
  FormSelect,
  SmallButton,
  StatusText,
} from '../../analytics/styled-components/analytics.styles';
import { fetchProfile, updateProfile } from '../apiService/profileApi';
import { ProfileFieldsGrid } from '../styled-components/profile.styles';
import {
  ACTIVITY_OPTIONS,
  SEX_OPTIONS,
  type ActivityLevel,
  type Sex,
  type UserProfile,
} from '../utils/types';
import { DailyTargetsRow } from './DailyTargetsRow';

type ProfileFormProps = {
  onUnauthorized: () => void;
};

const SEX_LABEL_KEYS: Record<Sex, 'sexMale' | 'sexFemale'> = {
  male: 'sexMale',
  female: 'sexFemale',
};

const ACTIVITY_LABEL_KEYS: Record<
  ActivityLevel,
  'actSedentary' | 'actLight' | 'actModerate' | 'actActive' | 'actVeryActive'
> = {
  sedentary: 'actSedentary',
  light: 'actLight',
  moderate: 'actModerate',
  active: 'actActive',
  very_active: 'actVeryActive',
};

export function ProfileForm({ onUnauthorized }: ProfileFormProps) {
  const { t } = useI18n();
  const [weight, setWeight] = useState('');
  const [height, setHeight] = useState('');
  const [age, setAge] = useState('');
  const [sex, setSex] = useState<Sex | ''>('');
  const [activity, setActivity] = useState<ActivityLevel | ''>('');
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const data = await fetchProfile();
        if (cancelled) {
          return;
        }
        setProfile(data);
        setWeight(data.weight_kg != null ? String(data.weight_kg) : '');
        setHeight(data.height_cm != null ? String(data.height_cm) : '');
        setAge(data.age != null ? String(data.age) : '');
        setSex(data.sex || '');
        setActivity(data.activity_level || '');
      } catch (err) {
        if (cancelled) {
          return;
        }
        const message = err instanceof Error ? err.message : t('profileLoadFailed');
        setError(message);
        if (message.includes('unauthorized')) {
          onUnauthorized();
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [onUnauthorized, t]);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);

    const weightKg = Number(weight);
    const heightCm = Number(height);
    const ageValue = Number(age);

    if (!sex || !activity) {
      setError(t('needSexActivity'));
      return;
    }
    if (!Number.isFinite(weightKg) || !Number.isFinite(heightCm) || !Number.isFinite(ageValue)) {
      setError(t('checkMetrics'));
      return;
    }

    setSaving(true);
    try {
      const updated = await updateProfile({
        weight_kg: weightKg,
        height_cm: heightCm,
        age: Math.round(ageValue),
        sex,
        activity_level: activity,
      });
      setProfile(updated);
    } catch (err) {
      const message = err instanceof Error ? err.message : t('profileSaveFailed');
      setError(message);
      if (message.includes('unauthorized')) {
        onUnauthorized();
      }
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return <StatusText>{t('profileLoading')}</StatusText>;
  }

  return (
    <>
      <FormPanel as="form" onSubmit={handleSubmit}>
        <ProfileFieldsGrid>
          <FormField>
            {t('weightKg')}
            <FormInput
              type="number"
              min={1}
              max={499}
              step="0.1"
              value={weight}
              onChange={(e) => setWeight(e.target.value)}
              required
            />
          </FormField>
          <FormField>
            {t('heightCm')}
            <FormInput
              type="number"
              min={1}
              max={299}
              step="0.1"
              value={height}
              onChange={(e) => setHeight(e.target.value)}
              required
            />
          </FormField>
          <FormField>
            {t('age')}
            <FormInput
              type="number"
              min={10}
              max={120}
              step="1"
              value={age}
              onChange={(e) => setAge(e.target.value)}
              required
            />
          </FormField>
          <FormField>
            {t('sex')}
            <FormSelect value={sex} onChange={(e) => setSex(e.target.value as Sex | '')} required>
              <option value="">{t('choose')}</option>
              {SEX_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {t(SEX_LABEL_KEYS[option.value])}
                </option>
              ))}
            </FormSelect>
          </FormField>
          <FormField>
            {t('activity')}
            <FormSelect
              value={activity}
              onChange={(e) => setActivity(e.target.value as ActivityLevel | '')}
              required
            >
              <option value="">{t('choose')}</option>
              {ACTIVITY_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {t(ACTIVITY_LABEL_KEYS[option.value])}
                </option>
              ))}
            </FormSelect>
          </FormField>
        </ProfileFieldsGrid>
        {error ? <StatusText $error>{error}</StatusText> : null}
        <ActionRow>
          <SmallButton type="submit" disabled={saving}>
            {saving ? t('saving') : t('calculateNorm')}
          </SmallButton>
        </ActionRow>
      </FormPanel>
      {profile?.daily_targets ? <DailyTargetsRow targets={profile.daily_targets} /> : null}
    </>
  );
}
