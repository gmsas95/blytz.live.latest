import 'package:blytz_flutter_app/core/constants/route_constants.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class ModernBottomNavigation extends StatelessWidget {
  const ModernBottomNavigation({required this.selectedIndex, super.key});

  final int selectedIndex;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    
    return NavigationBar(
      selectedIndex: selectedIndex,
      onDestinationSelected: (index) {
        switch (index) {
          case 0:
            context.push(RouteConstants.home);
          case 1:
            context.push(RouteConstants.discovery);
          case 2:
            context.push(RouteConstants.createStream);
          case 3:
            context.push(RouteConstants.community);
          case 4:
            context.push(RouteConstants.profile);
        }
      },
      backgroundColor: colorScheme.surface,
      surfaceTintColor: colorScheme.surfaceTint,
      indicatorColor: colorScheme.primary,
      labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
      destinations: const [
        NavigationDestination(
          icon: Icon(Icons.home_outlined),
          selectedIcon: Icon(Icons.home),
          label: 'Home',
        ),
        NavigationDestination(
          icon: Icon(Icons.explore_outlined),
          selectedIcon: Icon(Icons.explore),
          label: 'Discover',
        ),
        NavigationDestination(
          icon: Icon(Icons.live_tv_outlined),
          selectedIcon: Icon(Icons.live_tv),
          label: 'Go Live',
        ),
        NavigationDestination(
          icon: Icon(Icons.people_outline),
          selectedIcon: Icon(Icons.people),
          label: 'Community',
        ),
        NavigationDestination(
          icon: Icon(Icons.person_outline),
          selectedIcon: Icon(Icons.person),
          label: 'Profile',
        ),
      ],
    );
  }
}

// Backward compatibility alias
class BottomNavigation extends ModernBottomNavigation {
  const BottomNavigation({required super.selectedIndex, super.key});
}