<p align="center">
  <img src="https://raw.githubusercontent.com/futureharmony/storagebrowser/refs/heads/main/misc/static/Banner.png" width="550"/>
</p>


# Forked From FileBrowser.  And add **s3 storage support**!
This fork adds S3 storage support, allowing you to browse and manage files in S3-compatible object storage.
``` 
File Browser provides a file managing interface within a specified directory and it can be used to upload, delete, preview and edit your files. It is a create-your-own-cloud-kind of software where you can just install it on your server, direct it to a path and access your files through a nice web interface.
```


### Configuration

Add S3 settings to your configuration file (e.g., `.filebrowser.json`):

```json
{
  "port": "8080",
  "baseURL": "",
  "address": "127.0.0.1",
  "log": "stdout",
  "database": "filebrowser.db",
  "root": "/srv",
  "storageType": "s3",
  "s3Endpoint": "https://xxx.xx",
  "s3AccessKey": "",
  "s3SecretKey": "",
  "s3Region": "us-east-1"
}
```

### Command Line Usage

Start FileBrowser with S3 storage:

```bash
./filebrowser \
  --storage-type s3 \
  --s3-region us-east-1 \
  --s3-endpoint s3.amazonaws.com \
  --s3-access-key {{AK}} \
  --s3-secret-key {{SK}}
```

### Environment Variables

You can also use environment variables (FB_ prefix with underscore separators):

```bash
export FB_STORAGE_TYPE=s3
export FB_S3_REGION=us-east-1
export FB_S3_ENDPOINT=s3.amazonaws.com
export FB_S3_ACCESS_KEY=your-access-key
export FB_S3_SECRET_KEY=your-secret-key
```

### Bucket Switching

When using S3 storage, a bucket selector appears in the header, allowing you to:
- View current bucket name
- Switch between available buckets in your account
- Bucket selection persists across sessions

## License

[Apache License 2.0](LICENSE) © File Browser Contributors
