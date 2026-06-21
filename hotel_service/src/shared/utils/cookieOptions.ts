export const cookieOption = {
    httpOnly: true,
    secure: true,
    sameSite: 'strict',
} as const;

export const accessTokenCookieOptions = {
    ...cookieOption,
    maxAge: 3 * 24 * 60 * 60 * 1000, // 3 days (matches refresh token)
} as const;

export const refreshTokenCookieOptions = {
    ...cookieOption,
    maxAge: 3 * 24 * 60 * 60 * 1000, // 3 days
} as const;
