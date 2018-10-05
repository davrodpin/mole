/*
Package tunnel provides APIs to create SSH tunnels to perform local port
forwarding, leveraging the SSH configuration file (e.g. $HOME/.ssh/config) to
find specific attributes of the target ssh server like user name, port, host
name and key when not provided.

SSH Config File Support

The module looks for the ssh config file stored on $HOME/.ssh/config only.
There is no fallback support to try to use /etc/ssh/config.

The current API supports the following ssh config file options:

	Host
  Hostname
  User
  Port
  IdentityKey

For more information about SSH Local Port Forwarding, please visit:
https://www.ssh.com/ssh/tunneling/example#sec-Local-Forwarding

For more information about SSH Config File, please visit:
https://www.ssh.com/ssh/config/
*/
package tunnel
