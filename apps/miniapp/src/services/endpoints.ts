import { createBuocVangClient } from '@buocvang/api-client';
import { baseURL, getAccessToken, refreshAccessToken } from './api';

export type {
  Challenge,
  ChallengeLeaderboardEntry,
  CreateChallengePayload,
  LedgerEntry,
  LeaderboardEntry,
  ReferralStats,
  StepChunk,
  User,
  VoucherItem,
} from '@buocvang/api-client';

const client = createBuocVangClient({
  baseURL,
  getAccessToken,
  refreshAccessToken,
  clientName: 'zmp/0.1.0',
});

export const loginZalo = client.loginZalo;
export const signOut = client.signOut;

export const getMe = client.getMe;
export const patchMe = client.patchMe;
export const postAttribution = client.postAttribution;
export const checkHandle = client.checkHandle;

export const ingestSteps = client.ingestSteps;
export const getStepsToday = client.getStepsToday;
export const getStepsHistory = client.getStepsHistory;

export const listChallenges = client.listChallenges;
export const getChallenge = client.getChallenge;
export const createChallenge = client.createChallenge;
export const joinChallenge = client.joinChallenge;
export const challengeLeaderboard = client.challengeLeaderboard;
export const globalLeaderboard = client.globalLeaderboard;

export const getWalletBalance = client.getWalletBalance;
export const getLedger = client.getLedger;
export const listVouchers = client.listVouchers;
export const myVouchers = client.myVouchers;
export const redeemVoucher = client.redeemVoucher;

export const myReferral = client.myReferral;
export const trackReferral = client.trackReferral;

export const stravaAuthURL = client.stravaAuthURL;
export const stravaCallback = client.stravaCallback;
