import fs from 'fs/promises';
import path from 'path';
import handlebars from 'handlebars';
import { InternalServerError } from '../../shared/errors/app.error.js';
import { logger } from '../logger/index.js';

// Absolute path to the directory where all .hbs templates live
const TEMPLATES_DIR = path.join(process.cwd(), 'src', 'shared', 'template');

// In-memory cache of compiled templates, keyed by template name (avoids re-reading + re-compiling on every call)
const templateCache = new Map<string, HandlebarsTemplateDelegate>();

export async function compileTemplate(
    templateName: string,
    context: Record<string, unknown>
): Promise<string> {
    logger.info('Compiling template');
    // Strip any directory components from the input — only keep the file name itself
    const safeName = path.basename(templateName);

    // Build the full path to the .hbs file inside TEMPLATES_DIR
    const filePath = path.join(TEMPLATES_DIR, `${safeName}.hbs`);

    // Guard: if the resolved path somehow escapes TEMPLATES_DIR, reject it (path traversal protection)
    if (!filePath.startsWith(TEMPLATES_DIR)) {
        logger.error(`Invalid template name: "${templateName}"`);
        throw new InternalServerError(
            `Invalid template name: "${templateName}"`
        );
    }

    // Check the cache first to avoid disk I/O + recompilation
    let compiled = templateCache.get(safeName);

    // Cache miss — need to read and compile the template
    if (!compiled) {
        let source: string;
        try {
            // Read the raw .hbs file contents from disk (non-blocking)
            source = await fs.readFile(filePath, 'utf-8');
        } catch (err) {
            // File doesn't exist or isn't readable — surface a clear error
            logger.error(`Template "${safeName}" not found at ${filePath}`);
            throw new InternalServerError(
                `Template "${safeName}" not found at ${filePath}`
            );
        }

        // Compile the raw template string into a reusable Handlebars function
        compiled = handlebars.compile(source);

        // Store the compiled function in the cache for future calls
        templateCache.set(safeName, compiled);
    }

    logger.info('Template compiled successfully');
    // Render the compiled template with the provided context data and return the final HTML/string
    return compiled(context);
}

// Optional: useful in dev/hot-reload scenarios or tests
export function clearTemplateCache(): void {
    // Wipe the cache — forces next compileTemplate call to re-read from disk
    templateCache.clear();
}
