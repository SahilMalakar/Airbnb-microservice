import { z } from 'zod';

export const idSchema = z.object({
    id: z.coerce
        .number('id must be a number')
        .int('id must be an integer')
        .positive('id must be positive'),
});

export type IdSchemaDto = z.infer<typeof idSchema>;
