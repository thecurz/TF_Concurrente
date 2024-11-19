# Dockerfile para el servicio de la UI

# Usar una imagen oficial de nginx como base para servir archivos estáticos
FROM nginx:alpine

# Copiar el contenido de la carpeta UI al directorio de Nginx
COPY . /usr/share/nginx/html

# Exponer el puerto 80 para acceder a la UI
EXPOSE 80
