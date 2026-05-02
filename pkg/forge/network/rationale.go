package network

const whyMAC = `About MAC address generation:
  Random locally-administered unicast MAC address. The second-least-
  significant bit of the first octet is set (locally administered)
  and the least-significant bit is cleared (unicast). This ensures
  the address won't conflict with any hardware manufacturer's
  assigned addresses.

Why use it:
  VMs, containers, test environments, and network lab setups that
  need unique MAC addresses without risk of colliding with real
  hardware OUIs.

Note:
  Locally-administered addresses are recognized by all modern
  networking stacks. They will not appear in OUI lookup databases.

See also: smedje recommend vpn-key
`

func (m *MAC) Why() string { return whyMAC }
