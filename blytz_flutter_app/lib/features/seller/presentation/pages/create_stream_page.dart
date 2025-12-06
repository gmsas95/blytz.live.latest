import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';

class CreateStreamPage extends ConsumerStatefulWidget {
  const CreateStreamPage({super.key});

  @override
  ConsumerState<CreateStreamPage> createState() => _CreateStreamPageState();
}

class _CreateStreamPageState extends ConsumerState<CreateStreamPage> {
  final _formKey = GlobalKey<FormState>();
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _startingPriceController = TextEditingController();
  final _streamTitleController = TextEditingController();

  String? _selectedCategory;
  final List<String> _selectedProducts = [];
  final List<XFile> _selectedImages = [];
  DateTime? _scheduledTime;
  bool _isScheduled = false;

  final List<String> _categories = [
    'Electronics',
    'Fashion',
    'Collectibles',
    'Home & Garden',
    'Sports',
    'Toys & Games',
    'Books',
    'Art',
    'Jewelry',
    'Vintage',
  ];

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    _startingPriceController.dispose();
    _streamTitleController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Create Live Stream'),
        backgroundColor: Theme.of(context).primaryColor,
        foregroundColor: Colors.white,
        actions: [
          TextButton(
            onPressed: _createStream,
            child: const Text(
              'Go Live',
              style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
            ),
          ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Stream Information
              Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Text(
                  'Stream Information',
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),

              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      TextFormField(
                        controller: _streamTitleController,
                        decoration: const InputDecoration(
                          labelText: 'Stream Title',
                          border: OutlineInputBorder(),
                        ),
                        validator: (value) {
                          if (value == null || value.isEmpty) {
                            return 'Please enter a stream title';
                          }
                          return null;
                        },
                      ),
                      const SizedBox(height: 16),

                      TextFormField(
                        controller: _descriptionController,
                        maxLines: 3,
                        decoration: const InputDecoration(
                          labelText: 'Stream Description',
                          border: OutlineInputBorder(),
                        ),
                        validator: (value) {
                          if (value == null || value.isEmpty) {
                            return 'Please enter a stream description';
                          }
                          return null;
                        },
                      ),
                      const SizedBox(height: 16),

                      // Category Selection
                      const Padding(
                        padding: EdgeInsets.symmetric(vertical: 4.0),
                        child: Text(
                          'Category',
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),

                      DropdownButtonFormField<String>(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        borderRadius: BorderRadius.circular(8),
                        decoration: const InputDecoration(
                          border: OutlineInputBorder(),
                        ),
                        value: _selectedCategory,
                        hint: const Text('Select Category'),
                        items: _categories.map((category) {
                          return DropdownMenuItem(
                            value: category,
                            child: Text(category),
                          );
                        }).toList(),
                        onChanged: (value) {
                          setState(() {
                            _selectedCategory = value;
                          });
                        },
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 24),

              // Featured Product
              Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Text(
                  'Featured Product',
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),

              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      TextFormField(
                        controller: _titleController,
                        decoration: const InputDecoration(
                          labelText: 'Product Title',
                          border: OutlineInputBorder(),
                        ),
                        validator: (value) {
                          if (value == null || value.isEmpty) {
                            return 'Please enter a product title';
                          }
                          return null;
                        },
                      ),
                      const SizedBox(height: 16),

                      TextFormField(
                        controller: _startingPriceController,
                        keyboardType: const TextInputType.numberWithOptions(decimal: true),
                        decoration: const InputDecoration(
                          labelText: r'Starting Price ($)',
                          border: OutlineInputBorder(),
                        ),
                        validator: (value) {
                          if (value == null || value.isEmpty) {
                            return 'Please enter a starting price';
                          }
                          final price = double.tryParse(value);
                          if (price == null || price <= 0) {
                            return 'Please enter a valid price';
                          }
                          return null;
                        },
                      ),
                      const SizedBox(height: 16),

                      // Image Upload
                      const Padding(
                        padding: EdgeInsets.symmetric(vertical: 4.0),
                        child: Text(
                          'Product Images',
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),

                      _buildImageUpload(),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 24),

              // Scheduling Options
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 8.0),
                child: Text(
                  'Scheduling',
                  style: TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),

              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Checkbox(
                            value: _isScheduled,
                            onChanged: (value) {
                              setState(() {
                                _isScheduled = value ?? false;
                                if (!_isScheduled) {
                                  _scheduledTime = null;
                                }
                              });
                            },
                          ),
                          const SizedBox(width: 12),
                          const Text(
                            'Schedule for later',
                            style: TextStyle(
                              fontSize: 18,
                            ),
                          ),
                        ],
                      ),

                      if (_isScheduled) ...[
                        const SizedBox(height: 16),
                        FilledButton(
                          onPressed: _selectDateTime,
                          child: Text(
                            _scheduledTime != null
                                ? 'Scheduled: ${_scheduledTime!}'
                                : 'Select Date & Time',
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 24),

              // Stream Settings
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 8.0),
                child: Text(
                  'Stream Settings',
                  style: TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),

              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    children: [
                      _buildSettingTile(
                        'Enable Chat',
                        'Allow viewers to chat during stream',
                        true,
                        (value) {},
                      ),
                      const Divider(),
                      _buildSettingTile(
                        'Record Stream',
                        'Save recording for later viewing',
                        false,
                        (value) {},
                      ),
                      const Divider(),
                      _buildSettingTile(
                        'Send Notifications',
                        'Notify followers when stream starts',
                        true,
                        (value) {},
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 32),

              // Preview Section
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 8.0),
                child: Text(
                  'Preview',
                  style: TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),

              Card(
                color: Colors.grey[100],
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    children: [
                      // Thumbnail preview
                      Container(
                        height: 180,
                        decoration: BoxDecoration(
                          color: Colors.grey[300],
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: _selectedImages.isNotEmpty
                            ? ClipRRect(
                                borderRadius: BorderRadius.circular(8),
                                child: Image.network(
                                  _selectedImages.first.path,
                                  fit: BoxFit.cover,
                                  errorBuilder: (context, error, stackTrace) {
                                    return const Center(
                                      child: Icon(Icons.image, color: Colors.grey, size: 48),
                                    );
                                  },
                                ),
                              )
                            : const Center(
                                child: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    Icon(Icons.image, color: Colors.grey, size: 48),
                                    SizedBox(height: 8),
                                    Text('Stream Thumbnail', style: TextStyle(color: Colors.grey)),
                                  ],
                                ),
                              ),
                      ),

                      const SizedBox(height: 16),

                      // Stream info preview
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          if (_streamTitleController.text.isNotEmpty)
                            Text(
                              _streamTitleController.text,
                              style: const TextStyle(
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                              ),
                            )
                          else
                            Text(
                              'Your Stream Title',
                              style: TextStyle(
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                                color: Colors.grey[600] ?? Colors.grey,
                              ),
                            ),

                          const SizedBox(height: 4),

                          if (_descriptionController.text.isNotEmpty)
                            Text(
                              _descriptionController.text,
                              style: TextStyle(
                                fontSize: 14,
                                color: Colors.grey[600] ?? Colors.grey,
                              ),
                            )
                          else
                            Text(
                              'Stream description will appear here...',
                              style: TextStyle(
                                fontSize: 14,
                                color: Colors.grey[600] ?? Colors.grey,
                              ),
                            ),

                          const SizedBox(height: 8),

                          Row(
                            children: [
                              if (_selectedCategory != null) ...[
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                  decoration: BoxDecoration(
                                    color: Theme.of(context).primaryColor,
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  child: Text(
                                    _selectedCategory!,
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 14,
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 8),
                              ],
                              const Icon(Icons.visibility, color: Colors.grey, size: 16),
                              const SizedBox(width: 4),
                              Text(
                                '0 viewers',
                                style: TextStyle(
                                  color: Colors.grey[600] ?? Colors.grey,
                                  fontSize: 12,
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 32),

              // Create Stream Button
              FilledButton(
                onPressed: _createStream,
                style: FilledButton.styleFrom(
                  backgroundColor: Colors.red,
                  minimumSize: const Size(double.infinity, 48),
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.live_tv, color: Colors.white),
                    const SizedBox(width: 8),
                    Text(
                      _isScheduled ? 'Schedule Stream' : 'Go Live Now',
                      style: const TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildImageUpload() {
    return Column(
      children: [
        // Selected images
        if (_selectedImages.isNotEmpty) ...[
          SizedBox(
            height: 100,
            child: ListView.builder(
              scrollDirection: Axis.horizontal,
              itemCount: _selectedImages.length,
              itemBuilder: (context, index) {
                final image = _selectedImages[index];
                return Container(
                  width: 100,
                  margin: const EdgeInsets.only(right: 8),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.grey[300]!),
                  ),
                  child: Stack(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: Image.network(
                          image.path,
                          width: 100,
                          height: 100,
                          fit: BoxFit.cover,
                          errorBuilder: (context, error, stackTrace) {
                            return Container(
                              width: 100,
                              height: 100,
                              color: Colors.grey[200],
                              child: const Icon(Icons.image, color: Colors.grey),
                            );
                          },
                        ),
                      ),
                      Positioned(
                        top: 4,
                        right: 4,
                        child: GestureDetector(
                          onTap: () {
                            setState(() {
                              _selectedImages.removeAt(index);
                            });
                          },
                          child: Container(
                            padding: const EdgeInsets.all(4),
                            decoration: const BoxDecoration(
                              color: Colors.red,
                              shape: BoxShape.circle,
                            ),
                            child: const Icon(Icons.close, color: Colors.white, size: 16),
                          ),
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),
          const SizedBox(height: 12),
        ],

        // Add image button
        OutlinedButton.icon(
          onPressed: _addImage,
          icon: const Icon(Icons.add_photo_alternate),
          label: const Text('Add Image'),
        ),
      ],
    );
  }

  Widget _buildSettingTile(
    String title,
    String subtitle,
    bool value,
    Function(bool) onChanged,
  ) {
    return ListTile(
      contentPadding: EdgeInsets.zero,
      title: Text(title),
      subtitle: Text(
        subtitle,
        style: TextStyle(
          fontSize: 12,
          color: Colors.grey[600] ?? Colors.grey,
        ),
      ),
      trailing: Switch(
        value: value,
        onChanged: onChanged,
      ),
    );
  }

  Future<void> _addImage() async {
    final picker = ImagePicker();
    final images = await picker.pickMultiImage();
    if (images.isNotEmpty) {
      setState(() {
        _selectedImages.addAll(images.take(5 - _selectedImages.length)); // Max 5 images
      });
    }
  }

  Future<void> _selectDateTime() async {
    final date = await showDatePicker(
      context: context,
      initialDate: DateTime.now().add(const Duration(hours: 1)),
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 30)),
    );

    if (date != null) {
      final time = await showTimePicker(
        context: context,
        initialTime: TimeOfDay.now(),
      );

      if (time != null) {
        setState(() {
          _scheduledTime = DateTime(
            date.year,
            date.month,
            date.day,
            time.hour,
            time.minute,
          );
        });
      }
    }
  }

  void _createStream() {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    if (_selectedCategory == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please select a category'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    // Show success message
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(_isScheduled ? 'Stream scheduled successfully!' : 'Going live now!'),
        backgroundColor: Colors.green,
      ),
    );

    // Navigate back or to live stream page
    Navigator.pop(context);
  }
}