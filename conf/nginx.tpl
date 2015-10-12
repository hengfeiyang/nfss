server
{
    listen       80;
    server_name  [DOMAIN] [ALIAS];
    index index.shtml index.html index.htm index.php;
    root  [ROOT];
	set siteid [SITEID];
    location ~ /\.ht
    {
        deny all;
    }
    location ~ .*\.(php|php5)?$
    {
        fastcgi_pass  127.0.0.1:9000;
        fastcgi_index index.php;
        include fastcgi.conf;
    }
	# [CONNECTIONS] [BANDWIDTH]
    access_log [LOG] access;
}
