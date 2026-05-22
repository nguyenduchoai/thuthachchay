import { Routes, Route, Navigate } from 'react-router-dom';
import { Suspense, lazy } from 'react';
import { BottomNav } from './components/BottomNav';
import { useTranslation } from 'react-i18next';

const Splash = lazy(() => import('./pages/Splash'));
const Welcome = lazy(() => import('./pages/Welcome'));
const OnboardingHow = lazy(() => import('./pages/onboarding/How'));
const OnboardingSource = lazy(() => import('./pages/onboarding/Source'));
const OnboardingGoal = lazy(() => import('./pages/onboarding/Goal'));
const OnboardingUsername = lazy(() => import('./pages/onboarding/Username'));
const OnboardingLeaderboard = lazy(() => import('./pages/onboarding/LeaderboardPreview'));
const OnboardingStrava = lazy(() => import('./pages/onboarding/Strava'));
const OnboardingNotify = lazy(() => import('./pages/onboarding/Notify'));
const SignIn = lazy(() => import('./pages/auth/SignIn'));
const Home = lazy(() => import('./pages/Home'));
const Discover = lazy(() => import('./pages/Discover'));
const ChallengeDetail = lazy(() => import('./pages/ChallengeDetail'));
const Create = lazy(() => import('./pages/Create'));
const CreateNew = lazy(() => import('./pages/CreateNew'));
const Wallet = lazy(() => import('./pages/Wallet'));
const Profile = lazy(() => import('./pages/Profile'));
const ProfileSettings = lazy(() => import('./pages/ProfileSettings'));
const ProfileEdit = lazy(() => import('./pages/ProfileEdit'));
const Checkout = lazy(() => import('./pages/Checkout'));
const Invite = lazy(() => import('./pages/Invite'));

function PageLoader() {
  const { t } = useTranslation();
  return (
    <div className="page-loader">
      <div className="spinner" aria-label={t('common.loading')} />
    </div>
  );
}

export default function App() {
  return (
    <div className="app-shell">
      <Suspense fallback={<PageLoader />}>
        <Routes>
          <Route path="/" element={<Splash />} />
          <Route path="/welcome" element={<Welcome />} />
          <Route path="/onboarding/how" element={<OnboardingHow />} />
          <Route path="/onboarding/source" element={<OnboardingSource />} />
          <Route path="/onboarding/goal" element={<OnboardingGoal />} />
          <Route path="/onboarding/username" element={<OnboardingUsername />} />
          <Route path="/onboarding/leaderboard" element={<OnboardingLeaderboard />} />
          <Route path="/onboarding/strava" element={<OnboardingStrava />} />
          <Route path="/onboarding/notify" element={<OnboardingNotify />} />
          <Route path="/auth/sign-in" element={<SignIn />} />
          <Route path="/home" element={<Home />} />
          <Route path="/discover" element={<Discover />} />
          <Route path="/challenges/:id" element={<ChallengeDetail />} />
          <Route path="/create" element={<Create />} />
          <Route path="/create/new" element={<CreateNew />} />
          <Route path="/wallet" element={<Wallet />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/profile/settings" element={<ProfileSettings />} />
          <Route path="/profile/edit" element={<ProfileEdit />} />
          <Route path="/checkout/:tx" element={<Checkout />} />
          <Route path="/invite" element={<Invite />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Suspense>
      <BottomNav />
    </div>
  );
}
