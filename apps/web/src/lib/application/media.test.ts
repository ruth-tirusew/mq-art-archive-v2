import { describe, expect, it } from 'vitest';
import { buildUploadForm, uploadRejection, validateImageFile } from './media';

describe('validateImageFile', () => {
  it('rejects unsupported image types', () => {
    expect(validateImageFile({ type: 'image/gif', size: 100 })).toMatch(/JPEG, PNG, or WebP/);
  });

  it('rejects files larger than 10 MB', () => {
    expect(validateImageFile({ type: 'image/jpeg', size: 10 * 1024 * 1024 + 1 })).toMatch(
      /10 MB/
    );
  });

  it('accepts supported images up to 10 MB', () => {
    expect(validateImageFile({ type: 'image/webp', size: 10 * 1024 * 1024 })).toBeNull();
  });
});

describe('buildUploadForm', () => {
  const signature = {
    timestamp: 1700000000,
    signature: 'sig',
    cloud_name: 'cloud',
    api_key: 'key',
    folder: 'mq',
    public_id: 'mq/abc',
    expire_at: '2026-01-01T00:00:00Z'
  };

  it('sends every field the server signed, or Cloudinary rejects the signature', () => {
    const form = buildUploadForm(new Blob(['x']), signature);
    // The server signs folder, overwrite, public_id and timestamp (see Cloudinary.SignUpload).
    expect(form.get('folder')).toBe('mq');
    expect(form.get('overwrite')).toBe('false');
    expect(form.get('public_id')).toBe('mq/abc');
    expect(form.get('timestamp')).toBe('1700000000');
    expect(form.get('signature')).toBe('sig');
    expect(form.get('api_key')).toBe('key');
    expect(form.get('file')).toBeInstanceOf(Blob);
  });

  it('sends nothing the server did not sign (extra fields would also break the signature)', () => {
    const keys = [...buildUploadForm(new Blob(['x']), signature).keys()].sort();
    expect(keys).toEqual(['api_key', 'file', 'folder', 'overwrite', 'public_id', 'signature', 'timestamp']);
  });
});

describe('uploadRejection', () => {
  it('passes on Cloudinary’s own reason', () => {
    const body = JSON.stringify({ error: { message: 'Invalid Signature abc. String to sign - folder=mq' } });
    expect(uploadRejection(401, body)).toBe('Cloudinary rejected the image upload: Invalid Signature abc. String to sign - folder=mq');
  });

  it('falls back to the status when there is no usable body', () => {
    expect(uploadRejection(500, '')).toBe('Cloudinary rejected the image upload (HTTP 500).');
    expect(uploadRejection(400, '<html>nope</html>')).toBe('Cloudinary rejected the image upload (HTTP 400).');
  });
});
