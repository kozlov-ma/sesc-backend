#!/bin/bash
set -e

if [ ! -f /var/lib/ldap/.bootstrapped ]; then
    echo ">>> Bootstrapping LDAP..."
    sleep 5
    ldapadd -x -D "cn=admin,dc=lyceum,dc=usu,dc=ru" -w "${LDAP_ADMIN_PASSWORD}" -f /ldif/bootstrap.ldif
    touch /var/lib/ldap/.bootstrapped
else
    echo ">>> LDAP already bootstrapped"
fi
